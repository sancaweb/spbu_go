package server

import (
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// Register custom template functions
	r.SetFuncMap(template.FuncMap{
		"formatIDR": func(n int64) string {
			// Format integer as Indonesian number: 1.500.000
			if n == 0 {
				return "0"
			}
			neg := ""
			if n < 0 {
				neg = "-"
				n = -n
			}
			s := fmt.Sprintf("%d", n)
			result := ""
			for i, c := range s {
				if i > 0 && (len(s)-i)%3 == 0 {
					result += "."
				}
				result += string(c)
			}
			return neg + result
		},
	})

	// Static files
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// Custom template loading to avoid Windows path issues
	loadTemplates(r)

	return r
}

func loadTemplates(r *gin.Engine) {
	// Custom template loading to handle nested directories on Windows
	// We combine multiple patterns
	var files []string

	// Explicitly list patterns using filepath.Join for OS-correct separators
	patterns := []string{
		"templates/*.html",       // Root templates (error.html)
		"templates/*/*.html",     // Subdir (includes/header.html, main/home.html)
		"templates/*/*/*.html",   // Nested (master/bbm/index.html)
		"templates/*/*/*/*.html", // Deep nested (master/keuangan/coa/index.html)
	}

	for _, pattern := range patterns {
		// Handle slash replacement for Windows if needed, though Glob usually handles it
		// But to be safe, we can manually construct paths or just rely on Glob
		// Since the manual globs failed before, let's use filepath.Glob which is OS aware
		// But we must correct the pattern separators for Windows
		cleanPattern := filepath.FromSlash(pattern)
		matches, err := filepath.Glob(cleanPattern)
		if err != nil {
			continue
		}
		files = append(files, matches...)
	}

	if len(files) > 0 {
		r.LoadHTMLFiles(files...)
	}
}
