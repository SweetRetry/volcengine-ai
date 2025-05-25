package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"status_code":  param.StatusCode,
			"latency":      param.Latency,
			"client_ip":    param.ClientIP,
			"method":       param.Method,
			"path":         param.Path,
			"error":        param.ErrorMessage,
			"body_size":    param.BodySize,
			"user_agent":   param.Request.UserAgent(),
		}).Info("HTTP请求")

		return ""
	})
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"error": recovered,
			"path":  c.Request.URL.Path,
			"method": c.Request.Method,
		}).Error("服务器内部错误")

		c.JSON(500, gin.H{
			"error":   "服务器内部错误",
			"message": "请稍后重试",
		})
	})
} 