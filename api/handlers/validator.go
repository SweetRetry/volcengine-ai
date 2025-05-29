package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError 自定义验证错误结构
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidateRequest 统一的请求验证函数
func ValidateRequest(c *gin.Context, req interface{}) []ValidationError {
	var errors []ValidationError

	if err := c.ShouldBindJSON(req); err != nil {
		// 处理JSON绑定错误
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				errors = append(errors, ValidationError{
					Field:   fieldError.Field(),
					Message: getValidationMessage(fieldError),
					Value:   fieldError.Value(),
				})
			}
		} else {
			// 处理JSON格式错误
			errors = append(errors, ValidationError{
				Field:   "request_body",
				Message: "请求体格式错误，请确保发送有效的JSON",
				Value:   nil,
			})
		}
	}

	return errors
}

// getValidationMessage 根据验证规则返回中文错误信息
func getValidationMessage(fe validator.FieldError) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s 是必填字段", field)
	case "email":
		return fmt.Sprintf("%s 格式不正确", field)
	case "min":
		return fmt.Sprintf("%s 长度不能少于 %s 个字符", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s 长度不能超过 %s 个字符", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s 长度必须为 %s 个字符", field, fe.Param())
	case "numeric":
		return fmt.Sprintf("%s 必须是数字", field)
	case "alpha":
		return fmt.Sprintf("%s 只能包含字母", field)
	case "alphanum":
		return fmt.Sprintf("%s 只能包含字母和数字", field)
	default:
		return fmt.Sprintf("%s 验证失败", field)
	}
}

// ResponseValidationError 返回验证错误响应
func ResponseValidationError(c *gin.Context, errors []ValidationError) {
	c.JSON(400, gin.H{
		"error":   "请求参数验证失败",
		"message": "请检查以下字段",
		"details": errors,
	})
}
