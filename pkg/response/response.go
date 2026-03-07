package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string, errors interface{}) {
	c.JSON(code, Response{
		Status:  false,
		Message: message,
		Errors:  errors,
	})
}
