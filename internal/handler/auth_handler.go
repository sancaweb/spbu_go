package handler

import (
	"net/http"
	"spbu_go/internal/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService}
}

func (h *AuthHandler) LoginView(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title": "Login",
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, err := h.authService.Login(username, password)
	if err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"Error": "Invalid credentials",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("role_code", user.Role.Code)
	session.Save()

	c.Redirect(http.StatusFound, "/dashboard")
}

func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
