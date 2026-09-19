package server

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestImportDataTemplateRenders(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	router := NewRouter()
	router.GET("/test-import-data", func(c *gin.Context) {
		c.HTML(http.StatusOK, "settings/import_data.html", map[string]any{
			"ConnectionJSON": template.JS(`{"driver":"mysql","host":"127.0.0.1","port":3306}`),
			"Title":          "Import Data",
			"ActiveMenu":     "import_data",
		})
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/test-import-data", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	body := response.Body.String()
	if !strings.Contains(body, "Import Data") || !strings.Contains(body, "Master BBM") || !strings.Contains(body, `"driver":"mysql"`) {
		t.Fatal("import data template did not render expected content")
	}
}
