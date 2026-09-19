package handler

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"spbu_go/internal/dto"
)

const importDataUploadMaxBytes int64 = 5 << 20

var legacyBBMHeaders = []string{
	"source_id",
	"name",
	"margin",
	"price",
	"stock",
	"reward_percent",
	"is_active",
}

func buildLegacyBBMXLSX(rows []dto.LegacyBBMRow) ([]byte, error) {
	data := make([][]string, 0, len(rows))
	for _, row := range rows {
		data = append(data, []string{
			strconv.FormatInt(row.SourceID, 10),
			row.Name,
			formatImportNumber(row.Margin),
			formatImportNumber(row.Price),
			formatImportNumber(row.Stock),
			formatImportNumber(row.RewardPercent),
			strconv.FormatBool(row.IsActive),
		})
	}
	return buildImportXLSX("BBM", legacyBBMHeaders, data)
}

func buildLegacyBBMTemplateXLSX() ([]byte, error) {
	return buildImportXLSX("BBM", legacyBBMHeaders, nil)
}

func parseLegacyBBMXLSX(content []byte) ([]dto.LegacyBBMRow, error) {
	if int64(len(content)) > importDataUploadMaxBytes {
		return nil, fmt.Errorf("ukuran file import maksimal 5 MB")
	}
	records, err := parseXLSXRecords(content)
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("file Excel belum memiliki data BBM")
	}

	headers := make(map[string]int, len(records[0]))
	for index, value := range records[0] {
		if key := normalizeSalesReportHeader(value); key != "" {
			headers[key] = index
		}
	}
	nameIndex := findHeaderIndex(headers, "name", "nama", "nama_bbm")
	if nameIndex < 0 {
		return nil, fmt.Errorf("kolom wajib name tidak ditemukan")
	}

	indexes := struct {
		sourceID, margin, price, stock, reward, active int
	}{
		sourceID: findHeaderIndex(headers, "source_id", "id_bbm", "idbbm"),
		margin:   findHeaderIndex(headers, "margin"),
		price:    findHeaderIndex(headers, "price", "harga", "harga_jual"),
		stock:    findHeaderIndex(headers, "stock", "stok", "stok_liter"),
		reward:   findHeaderIndex(headers, "reward_percent", "reward_persen", "reward"),
		active:   findHeaderIndex(headers, "is_active", "status", "aktif"),
	}

	rows := make([]dto.LegacyBBMRow, 0, len(records)-1)
	for lineNumber, record := range records[1:] {
		if isEmptyCSVRecord(record) {
			continue
		}
		lineNumber += 2
		name := csvValue(record, nameIndex)
		if name == "" {
			return nil, fmt.Errorf("baris %d memiliki nama BBM kosong", lineNumber)
		}

		sourceID, err := parseImportInt(csvValue(record, indexes.sourceID))
		if err != nil {
			return nil, fmt.Errorf("baris %d memiliki source_id tidak valid", lineNumber)
		}
		margin, err := parseImportNumber(csvValue(record, indexes.margin))
		if err != nil {
			return nil, fmt.Errorf("baris %d memiliki margin tidak valid", lineNumber)
		}
		price, err := parseImportNumber(csvValue(record, indexes.price))
		if err != nil {
			return nil, fmt.Errorf("baris %d memiliki price tidak valid", lineNumber)
		}
		stock, err := parseImportNumber(csvValue(record, indexes.stock))
		if err != nil {
			return nil, fmt.Errorf("baris %d memiliki stock tidak valid", lineNumber)
		}
		reward, err := parseImportNumber(csvValue(record, indexes.reward))
		if err != nil {
			return nil, fmt.Errorf("baris %d memiliki reward_percent tidak valid", lineNumber)
		}

		active := true
		if value := csvValue(record, indexes.active); value != "" {
			active, err = parseImportBool(value)
			if err != nil {
				return nil, fmt.Errorf("baris %d memiliki is_active tidak valid", lineNumber)
			}
		}
		rows = append(rows, dto.LegacyBBMRow{
			SourceID:      sourceID,
			Name:          name,
			Margin:        margin,
			Price:         price,
			Stock:         stock,
			RewardPercent: reward,
			IsActive:      active,
		})
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("file Excel belum memiliki data BBM")
	}
	return rows, nil
}

func parseImportNumber(value string) (float64, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	return parseSalesReportNumber(value)
}

func parseImportInt(value string) (int64, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	parsed, err := parseImportNumber(value)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed < math.MinInt64 || parsed > math.MaxInt64 || parsed != math.Trunc(parsed) {
		return 0, fmt.Errorf("bukan bilangan bulat")
	}
	return int64(parsed), nil
}

func parseImportBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "aktif", "active":
		return true, nil
	case "0", "false", "no", "n", "tidak aktif", "inactive":
		return false, nil
	default:
		return false, fmt.Errorf("gunakan true/false atau aktif/tidak aktif")
	}
}

func formatImportNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func buildImportXLSX(sheetName string, headers []string, rows [][]string) ([]byte, error) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	entries := []struct {
		name string
		data string
	}{
		{"[Content_Types].xml", importXLSXContentTypesXML},
		{"_rels/.rels", importXLSXRootRelationshipsXML},
		{"xl/workbook.xml", fmt.Sprintf(importXLSXWorkbookXML, xmlEscape(sheetName))},
		{"xl/_rels/workbook.xml.rels", importXLSXWorkbookRelationshipsXML},
		{"xl/styles.xml", importXLSXStylesXML},
		{"xl/worksheets/sheet1.xml", buildImportWorksheetXML(headers, rows)},
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

func buildImportWorksheetXML(headers []string, rows [][]string) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)

	writeRow := func(rowNumber int, values []string, header bool) {
		fmt.Fprintf(&builder, `<row r="%d">`, rowNumber)
		for column, value := range values {
			cellReference := excelColumnName(column) + strconv.Itoa(rowNumber)
			style := 0
			if header {
				style = 1
			}
			fmt.Fprintf(&builder, `<c r="%s" t="inlineStr" s="%d"><is><t xml:space="preserve">`, cellReference, style)
			_ = xml.EscapeText(&builder, []byte(value))
			builder.WriteString(`</t></is></c>`)
		}
		builder.WriteString(`</row>`)
	}

	writeRow(1, headers, true)
	for index, row := range rows {
		writeRow(index+2, row, false)
	}
	builder.WriteString(`</sheetData><autoFilter ref="A1:G`)
	builder.WriteString(strconv.Itoa(len(rows) + 1))
	builder.WriteString(`"/></worksheet>`)
	return builder.String()
}

func excelColumnName(index int) string {
	name := ""
	for index >= 0 {
		name = string(rune('A'+index%26)) + name
		index = index/26 - 1
	}
	return name
}

func xmlEscape(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}

const importXLSXContentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`

const importXLSXRootRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const importXLSXWorkbookXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets><sheet name="%s" sheetId="1" r:id="rId1"/></sheets>
</workbook>`

const importXLSXWorkbookRelationshipsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

const importXLSXStylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<fonts count="2"><font><sz val="11"/><color rgb="FF000000"/><name val="Calibri"/></font><font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Calibri"/></font></fonts>
<fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FF1F4E78"/><bgColor indexed="64"/></patternFill></fill></fills>
<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>
<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
</cellStyleXfs><cellXfs count="2"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="1" fillId="1" borderId="0" xfId="0" applyFont="1" applyFill="1"><alignment horizontal="center" vertical="center"/></xf></cellXfs>
<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`
