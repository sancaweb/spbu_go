package handler

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

const salesReportUploadMaxBytes int64 = 5 << 20

const (
	maxImportWorksheetRows    = 10000
	maxImportWorksheetColumns = 256
)

type salesReportRow struct {
	NozzleID         uint
	Description      string
	TotalisatorAwal  float64
	TotalisatorAkhir float64
}

// UploadSalesReport menerima laporan penjualan dalam Excel atau CSV. File
// hanya berisi data totalisator; piutang dan pengeluaran test tetap dikelola
// dari form.
func (h *PenjualanHandler) UploadSalesReport(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, salesReportUploadMaxBytes)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "File laporan penjualan wajib dipilih"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".xlsx" && ext != ".csv" && ext != ".txt" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Format file harus Excel (.xlsx). Gunakan tombol Unduh Template untuk format yang sesuai",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "File laporan tidak dapat dibaca"})
		return
	}
	defer file.Close()

	reportRows, err := parseSalesReportUpload(file, ext)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	availableJSON, err := h.buildNozzleRows(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal memuat daftar nozzle aktif: " + err.Error()})
		return
	}
	var availableRows []NozzleFormRow
	if err := json.Unmarshal([]byte(availableJSON), &availableRows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal menyiapkan daftar nozzle aktif"})
		return
	}

	rows, err := mergeSalesReportRows(availableRows, reportRows)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Laporan penjualan berhasil dimuat untuk %d nozzle aktif", len(rows)),
		"rows":    rows,
	})
}

// DownloadSalesReportTemplate membuat template Excel berdasarkan nozzle aktif
// agar pengguna tidak perlu menebak ID nozzle yang harus diisi.
func (h *PenjualanHandler) DownloadSalesReportTemplate(c *gin.Context) {
	rowsJSON, err := h.buildNozzleRows(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal memuat daftar nozzle aktif"})
		return
	}
	var rows []NozzleFormRow
	if err := json.Unmarshal([]byte(rowsJSON), &rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal membuat template laporan"})
		return
	}

	content, err := buildSalesReportXLSX(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal membuat template laporan"})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="template_laporan_penjualan.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", content)
}

func parseSalesReportUpload(reader io.Reader, extension string) ([]salesReportRow, error) {
	content, err := io.ReadAll(io.LimitReader(reader, salesReportUploadMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file laporan: %w", err)
	}
	if int64(len(content)) > salesReportUploadMaxBytes {
		return nil, fmt.Errorf("ukuran file laporan maksimal 5 MB")
	}

	if extension == ".xlsx" {
		return parseSalesReportXLSX(content)
	}
	return parseSalesReportCSV(bytes.NewReader(content))
}

func parseSalesReportCSV(reader io.Reader) ([]salesReportRow, error) {
	content, err := io.ReadAll(io.LimitReader(reader, salesReportUploadMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file laporan: %w", err)
	}
	if int64(len(content)) > salesReportUploadMaxBytes {
		return nil, fmt.Errorf("ukuran file laporan maksimal 5 MB")
	}
	if len(bytes.TrimSpace(content)) == 0 {
		return nil, fmt.Errorf("file laporan kosong")
	}

	content = bytes.TrimPrefix(content, []byte{0xEF, 0xBB, 0xBF})
	comma := ','
	firstLine := content
	if lineEnd := bytes.IndexByte(content, '\n'); lineEnd >= 0 {
		firstLine = content[:lineEnd]
	}
	if bytes.Count(firstLine, []byte(";")) > bytes.Count(firstLine, []byte(",")) {
		comma = ';'
	}

	csvReader := csv.NewReader(bytes.NewReader(content))
	csvReader.Comma = comma
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true
	csvReader.ReuseRecord = false

	header, err := csvReader.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("file laporan tidak memiliki header")
	}
	if err != nil {
		return nil, fmt.Errorf("header CSV tidak valid: %w", err)
	}

	headerIndexes := make(map[string]int, len(header))
	for index, value := range header {
		key := normalizeSalesReportHeader(value)
		if key != "" {
			headerIndexes[key] = index
		}
	}

	nozzleIDIndex := findHeaderIndex(headerIndexes, "nozzle_id", "id_nozzle", "nozzleid")
	nozzleIndex := findHeaderIndex(headerIndexes, "nozzle", "nomor_nozzle", "nozzle_number", "description")
	startIndex := findHeaderIndex(headerIndexes, "totalisator_awal", "totalisator_start", "awal")
	endIndex := findHeaderIndex(headerIndexes, "totalisator_akhir", "totalisator_end", "akhir")
	if startIndex < 0 || endIndex < 0 {
		return nil, fmt.Errorf("kolom wajib totalisator_awal dan totalisator_akhir tidak ditemukan")
	}
	if nozzleIDIndex < 0 && nozzleIndex < 0 {
		return nil, fmt.Errorf("kolom wajib nozzle_id atau nozzle tidak ditemukan")
	}

	var rows []salesReportRow
	for lineNumber := 2; ; lineNumber++ {
		record, readErr := csvReader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("baris %d tidak valid: %w", lineNumber, readErr)
		}
		if isEmptyCSVRecord(record) {
			continue
		}

		identifier := csvValue(record, nozzleIDIndex)
		description := csvValue(record, nozzleIndex)
		nozzleID := uint(0)
		if identifier != "" {
			parsedID, parseErr := strconv.ParseUint(strings.TrimSpace(identifier), 10, 32)
			if parseErr != nil || parsedID == 0 {
				return nil, fmt.Errorf("baris %d memiliki nozzle_id tidak valid", lineNumber)
			}
			nozzleID = uint(parsedID)
		}

		start, parseErr := parseSalesReportNumber(csvValue(record, startIndex))
		if parseErr != nil {
			return nil, fmt.Errorf("baris %d memiliki totalisator_awal tidak valid", lineNumber)
		}
		end, parseErr := parseSalesReportNumber(csvValue(record, endIndex))
		if parseErr != nil {
			return nil, fmt.Errorf("baris %d memiliki totalisator_akhir tidak valid", lineNumber)
		}
		if start < 0 || end < 0 || end < start {
			return nil, fmt.Errorf("baris %d memiliki totalisator akhir lebih kecil dari totalisator awal", lineNumber)
		}

		rows = append(rows, salesReportRow{
			NozzleID:         nozzleID,
			Description:      description,
			TotalisatorAwal:  start,
			TotalisatorAkhir: end,
		})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("file laporan tidak memiliki data nozzle")
	}
	return rows, nil
}

type xlsxWorkbookXML struct {
	Sheets struct {
		Items []xlsxSheetXML `xml:"sheet"`
	} `xml:"sheets"`
}

type xlsxSheetXML struct {
	Name string `xml:"name,attr"`
	ID   string `xml:"id,attr"`
}

type xlsxRelationshipsXML struct {
	Items []xlsxRelationshipXML `xml:"Relationship"`
}

type xlsxRelationshipXML struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type xlsxSharedStringsXML struct {
	Items []xlsxSharedStringItemXML `xml:"si"`
}

type xlsxSharedStringItemXML struct {
	Texts []string         `xml:"t"`
	Runs  []xlsxTextRunXML `xml:"r"`
}

type xlsxTextRunXML struct {
	Text string `xml:"t"`
}

type xlsxWorksheetXML struct {
	SheetData struct {
		Rows []xlsxRowXML `xml:"row"`
	} `xml:"sheetData"`
}

type xlsxRowXML struct {
	Cells []xlsxCellXML `xml:"c"`
}

type xlsxCellXML struct {
	Reference string `xml:"r,attr"`
	Type      string `xml:"t,attr"`
	Value     string `xml:"v"`
	Inline    struct {
		Texts []string         `xml:"t"`
		Runs  []xlsxTextRunXML `xml:"r"`
	} `xml:"is"`
}

func parseSalesReportXLSX(content []byte) ([]salesReportRow, error) {
	records, err := parseXLSXRecords(content)
	if err != nil {
		return nil, err
	}

	var csvBuffer bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuffer)
	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return nil, fmt.Errorf("data worksheet Excel tidak dapat diproses")
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return nil, fmt.Errorf("data worksheet Excel tidak dapat diproses")
	}
	return parseSalesReportCSV(bytes.NewReader(csvBuffer.Bytes()))
}

// parseXLSXRecords reads the first worksheet as rows of text. It is shared by
// import modules so every Excel upload follows the same bounded parser path.
func parseXLSXRecords(content []byte) ([][]string, error) {
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("file Excel tidak valid atau rusak")
	}

	workbookContent, err := readXLSXEntry(archive, "xl/workbook.xml")
	if err != nil {
		return nil, fmt.Errorf("struktur workbook Excel tidak ditemukan")
	}
	var workbook xlsxWorkbookXML
	if err := xml.Unmarshal(workbookContent, &workbook); err != nil {
		return nil, fmt.Errorf("workbook Excel tidak dapat dibaca")
	}

	sheetPath := "xl/worksheets/sheet1.xml"
	if len(workbook.Sheets.Items) > 0 && workbook.Sheets.Items[0].ID != "" {
		relationshipContent, relErr := readXLSXEntry(archive, "xl/_rels/workbook.xml.rels")
		if relErr == nil {
			var relationships xlsxRelationshipsXML
			if xml.Unmarshal(relationshipContent, &relationships) == nil {
				for _, sheet := range workbook.Sheets.Items {
					if sheet.ID != workbook.Sheets.Items[0].ID {
						continue
					}
					for _, relationship := range relationships.Items {
						if relationship.ID == sheet.ID {
							sheetPath = resolveXLSXTarget(relationship.Target)
							break
						}
					}
				}
			}
		}
	}

	sheetContent, err := readXLSXEntry(archive, sheetPath)
	if err != nil {
		return nil, fmt.Errorf("worksheet Excel tidak ditemukan")
	}
	var worksheet xlsxWorksheetXML
	if err := xml.Unmarshal(sheetContent, &worksheet); err != nil {
		return nil, fmt.Errorf("worksheet Excel tidak dapat dibaca")
	}

	sharedStrings := []string{}
	if sharedContent, sharedErr := readXLSXEntry(archive, "xl/sharedStrings.xml"); sharedErr == nil {
		var shared xlsxSharedStringsXML
		if xml.Unmarshal(sharedContent, &shared) == nil {
			sharedStrings = make([]string, len(shared.Items))
			for index, item := range shared.Items {
				sharedStrings[index] = strings.Join(item.Texts, "")
				if len(item.Runs) > 0 {
					var builder strings.Builder
					for _, run := range item.Runs {
						builder.WriteString(run.Text)
					}
					sharedStrings[index] = builder.String()
				}
			}
		}
	}

	records := make([][]string, 0, len(worksheet.SheetData.Rows))
	for _, row := range worksheet.SheetData.Rows {
		if len(records) >= maxImportWorksheetRows {
			return nil, fmt.Errorf("worksheet Excel terlalu banyak baris")
		}
		maxColumn := -1
		for _, cell := range row.Cells {
			column := xlsxCellColumn(cell.Reference)
			if column >= maxImportWorksheetColumns {
				return nil, fmt.Errorf("worksheet Excel memiliki terlalu banyak kolom")
			}
			if column > maxColumn {
				maxColumn = column
			}
		}
		if maxColumn < 0 {
			continue
		}
		record := make([]string, maxColumn+1)
		for _, cell := range row.Cells {
			column := xlsxCellColumn(cell.Reference)
			if column < 0 {
				continue
			}
			value := cell.Value
			switch cell.Type {
			case "s":
				index, parseErr := strconv.Atoi(value)
				if parseErr != nil || index < 0 || index >= len(sharedStrings) {
					return nil, fmt.Errorf("nilai teks pada worksheet Excel tidak valid")
				}
				value = sharedStrings[index]
			case "inlineStr":
				value = strings.Join(cell.Inline.Texts, "")
				if len(cell.Inline.Runs) > 0 {
					var builder strings.Builder
					for _, run := range cell.Inline.Runs {
						builder.WriteString(run.Text)
					}
					value = builder.String()
				}
			}
			record[column] = value
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("worksheet Excel tidak memiliki data")
	}
	return records, nil
}

func readXLSXEntry(archive *zip.Reader, name string) ([]byte, error) {
	for _, entry := range archive.File {
		if entry.Name != name {
			continue
		}
		reader, err := entry.Open()
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		content, err := io.ReadAll(io.LimitReader(reader, salesReportUploadMaxBytes+1))
		if err != nil {
			return nil, err
		}
		if int64(len(content)) > salesReportUploadMaxBytes {
			return nil, fmt.Errorf("worksheet Excel terlalu besar")
		}
		return content, nil
	}
	return nil, fmt.Errorf("entry %s tidak ditemukan", name)
}

func resolveXLSXTarget(target string) string {
	target = strings.TrimPrefix(target, "/")
	if strings.HasPrefix(target, "xl/") {
		return target
	}
	return "xl/" + strings.TrimPrefix(target, "./")
}

func xlsxCellColumn(reference string) int {
	column := 0
	for _, character := range strings.ToUpper(reference) {
		if character < 'A' || character > 'Z' {
			break
		}
		column = (column * 26) + int(character-'A'+1)
	}
	if column == 0 {
		return -1
	}
	return column - 1
}

func buildSalesReportXLSX(rows []NozzleFormRow) ([]byte, error) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)

	entries := []struct {
		name string
		data string
	}{
		{"[Content_Types].xml", xlsxContentTypesXML},
		{"_rels/.rels", xlsxRootRelationshipsXML},
		{"xl/workbook.xml", xlsxWorkbookXMLContent},
		{"xl/_rels/workbook.xml.rels", xlsxWorkbookRelationshipsXML},
		{"xl/styles.xml", xlsxStylesXML},
		{"xl/worksheets/sheet1.xml", buildSalesReportWorksheetXML(rows)},
	}
	for _, entry := range entries {
		writer, err := archive.Create(entry.name)
		if err != nil {
			return nil, err
		}
		if _, err := io.WriteString(writer, entry.data); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func buildSalesReportWorksheetXML(rows []NozzleFormRow) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`)
	fmt.Fprintf(&builder, `<dimension ref="A1:D%d"/>`, len(rows)+1)
	builder.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
	builder.WriteString(`<sheetFormatPr defaultRowHeight="18"/><cols>`)
	builder.WriteString(`<col min="1" max="1" width="12" customWidth="1"/><col min="2" max="2" width="18" customWidth="1"/><col min="3" max="4" width="22" customWidth="1"/>`)
	builder.WriteString(`</cols><sheetData>`)

	writeInlineStringCell := func(cellReference, value string, style int) {
		fmt.Fprintf(&builder, `<c r="%s" t="inlineStr" s="%d"><is><t xml:space="preserve">`, cellReference, style)
		_ = xml.EscapeText(&builder, []byte(value))
		builder.WriteString(`</t></is></c>`)
	}
	writeNumberCell := func(cellReference, value string, style int) {
		fmt.Fprintf(&builder, `<c r="%s" s="%d"><v>%s</v></c>`, cellReference, style, value)
	}

	builder.WriteString(`<row r="1" ht="22" customHeight="1">`)
	writeInlineStringCell("A1", "nozzle_id", 1)
	writeInlineStringCell("B1", "nozzle", 1)
	writeInlineStringCell("C1", "totalisator_awal", 1)
	writeInlineStringCell("D1", "totalisator_akhir", 1)
	builder.WriteString(`</row>`)
	for index, row := range rows {
		excelRow := index + 2
		fmt.Fprintf(&builder, `<row r="%d">`, excelRow)
		writeNumberCell("A"+strconv.Itoa(excelRow), strconv.FormatUint(uint64(row.NozzleID), 10), 0)
		writeInlineStringCell("B"+strconv.Itoa(excelRow), row.Description, 0)
		writeNumberCell("C"+strconv.Itoa(excelRow), "0", 2)
		writeNumberCell("D"+strconv.Itoa(excelRow), "0", 2)
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData><autoFilter ref="A1:D`)
	builder.WriteString(strconv.Itoa(len(rows) + 1))
	builder.WriteString(`"/><pageMargins left="0.25" right="0.25" top="0.5" bottom="0.5" header="0" footer="0"/></worksheet>`)
	return builder.String()
}

const xlsxContentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`

const xlsxRootRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const xlsxWorkbookXMLContent = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets><sheet name="Penjualan" sheetId="1" r:id="rId1"/></sheets>
</workbook>`

const xlsxWorkbookRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

const xlsxStylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<numFmts count="1"><numFmt numFmtId="164" formatCode="0.00"/></numFmts>
<fonts count="2"><font><sz val="11"/><color rgb="FF000000"/><name val="Calibri"/></font><font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Calibri"/></font></fonts>
<fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FF1F4E78"/><bgColor indexed="64"/></patternFill></fill></fills>
<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>
<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
<cellXfs count="3"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="1" fillId="1" borderId="0" xfId="0" applyFont="1" applyFill="1"><alignment horizontal="center" vertical="center"/></xf><xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0"/></cellXfs>
<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`

func mergeSalesReportRows(available []NozzleFormRow, uploaded []salesReportRow) ([]NozzleFormRow, error) {
	byID := make(map[uint]int, len(available))
	byDescription := make(map[string][]int, len(available))
	for index, row := range available {
		byID[row.NozzleID] = index
		description := normalizeSalesReportHeader(row.Description)
		if description != "" {
			byDescription[description] = append(byDescription[description], index)
		}
	}

	seen := make(map[int]struct{}, len(uploaded))
	for _, source := range uploaded {
		index := -1
		if source.NozzleID > 0 {
			var ok bool
			index, ok = byID[source.NozzleID]
			if !ok {
				return nil, fmt.Errorf("nozzle_id %d tidak terdaftar sebagai nozzle aktif", source.NozzleID)
			}
		} else {
			matches := byDescription[normalizeSalesReportHeader(source.Description)]
			if len(matches) != 1 {
				return nil, fmt.Errorf("nozzle %q tidak dapat dipetakan secara unik", source.Description)
			}
			index = matches[0]
		}
		if _, duplicate := seen[index]; duplicate {
			return nil, fmt.Errorf("data nozzle %d muncul lebih dari sekali", available[index].NozzleID)
		}
		seen[index] = struct{}{}

		available[index].TotalisatorAwal = source.TotalisatorAwal
		available[index].TotalisatorAkhir = source.TotalisatorAkhir
		available[index].JmlLiter = source.TotalisatorAkhir - source.TotalisatorAwal
		available[index].JmlRupiah = int64(math.Round(available[index].JmlLiter * float64(available[index].BBMPrice)))
	}

	if len(seen) != len(available) {
		missing := make([]string, 0, len(available)-len(seen))
		for index, row := range available {
			if _, ok := seen[index]; !ok {
				missing = append(missing, fmt.Sprintf("%d (%s)", row.NozzleID, row.Description))
			}
		}
		return nil, fmt.Errorf("laporan belum mencakup semua nozzle aktif; belum ada data: %s", strings.Join(missing, ", "))
	}

	return available, nil
}

func parseSalesReportNumber(value string) (float64, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "\ufeff"))
	value = strings.ReplaceAll(value, "Rp", "")
	value = strings.ReplaceAll(value, "rp", "")
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "'", "")
	if value == "" {
		return 0, fmt.Errorf("nilai kosong")
	}

	lastDot := strings.LastIndex(value, ".")
	lastComma := strings.LastIndex(value, ",")
	switch {
	case lastDot >= 0 && lastComma >= 0 && lastComma > lastDot:
		value = strings.ReplaceAll(value, ".", "")
		value = strings.Replace(value, ",", ".", 1)
	case lastDot >= 0 && lastComma >= 0:
		value = strings.ReplaceAll(value, ",", "")
	case lastComma >= 0:
		value = strings.Replace(value, ",", ".", 1)
	}

	return strconv.ParseFloat(value, 64)
}

func normalizeSalesReportHeader(value string) string {
	value = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "\ufeff")))
	var builder strings.Builder
	previousUnderscore := false
	for _, character := range value {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			builder.WriteRune(character)
			previousUnderscore = false
			continue
		}
		if !previousUnderscore {
			builder.WriteByte('_')
			previousUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}

func findHeaderIndex(headers map[string]int, candidates ...string) int {
	for _, candidate := range candidates {
		if index, ok := headers[candidate]; ok {
			return index
		}
	}
	return -1
}

func csvValue(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func isEmptyCSVRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
