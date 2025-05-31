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

// CreatedResponse 创建成功响应 (201)
func CreatedResponse(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// AcceptedResponse 已接受响应 (202)
func AcceptedResponse(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusAccepted, Response{
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

// BadRequestResponse 请求错误响应 (400)
func BadRequestResponse(c *gin.Context, error string, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   error,
		Message: message,
	})
}

// NotFoundResponse 未找到响应 (404)
func NotFoundResponse(c *gin.Context, error string, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error:   error,
		Message: message,
	})
}

// InternalServerErrorResponse 服务器内部错误响应 (500)
func InternalServerErrorResponse(c *gin.Context, error string, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   error,
		Message: message,
	})
}

// NotImplementedResponse 未实现响应 (501)
func NotImplementedResponse(c *gin.Context, error string, message string, data interface{}) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   error,
		"message": message,
		"data":    data,
	})
}

// TooManyRequestsResponse 请求过多响应 (429)
func TooManyRequestsResponse(c *gin.Context, error string, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
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
