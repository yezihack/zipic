package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a unified API response
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// Success returns a success response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// Error returns an error response
func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// BadRequest returns a bad request error
func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, msg)
}

// InternalError returns an internal server error
func InternalError(c *gin.Context, msg string) {
	Error(c, http.StatusInternalServerError, msg)
}