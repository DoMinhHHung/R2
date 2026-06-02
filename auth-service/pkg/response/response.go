package response

import (
	"errors"
	"net/http"

	"github.com/DoMinhHHung/auth-service/pkg/apperr"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool   `json:"success"`
	Data      any    `json:"data,omitempty"`
	Message   string `json:"message,omitempty"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

func OKMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Message:   message,
		RequestID: getRequestID(c),
	})
}

func Error(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, Response{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			RequestID: getRequestID(c),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Response{
		Success:   false,
		Code:      "INTERNAL_ERROR",
		Message:   "An internal error occurred",
		RequestID: getRequestID(c),
	})
}

func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success:   false,
		Code:      "VALIDATION_ERROR",
		Message:   message,
		RequestID: getRequestID(c),
	})
}

func getRequestID(c *gin.Context) string {
	rid, _ := c.Get("request_id")
	if s, ok := rid.(string); ok {
		return s
	}
	return ""
}
