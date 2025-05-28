package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, statusCode int, error string, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   error,
		Message: message,
	})
}

// ValidationErrorResponse 验证错误响应
func ValidationErrorResponse(c *gin.Context, details interface{}) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error":   "请求参数验证失败",
		"message": "请检查以下字段",
		"details": details,
	})
}
