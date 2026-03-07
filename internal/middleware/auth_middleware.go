package middleware

import (
	"net/http"
	"spbu_go/internal/repository"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		user, err := userRepo.FindByID(userID.(uint))
		if err != nil {
			c.Redirect(http.StatusFound, "/logout")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func GuestOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user_id")
		if user != nil {
			c.Redirect(http.StatusFound, "/dashboard")
			c.Abort()
			return
		}
		c.Next()
	}
}
