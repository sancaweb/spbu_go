package middleware

import (
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingsMiddleware injects site settings into the context for all templates
func SettingsMiddleware(settingService service.SettingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		favicon, _ := settingService.Get("favicon")
		if favicon == "" {
			favicon = "/static/favicon.png"
		}
		c.Set("favicon", favicon)

		// Store settings in context for template access
		c.Set("site_settings", map[string]string{
			"favicon": favicon,
		})
		c.Next()
	}
}
