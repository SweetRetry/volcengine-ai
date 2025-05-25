package middleware

import (
	"github.com/gin-gonic/gin"
)

// OptionsHandler 处理OPTIONS请求的中间件
func OptionsHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
}
