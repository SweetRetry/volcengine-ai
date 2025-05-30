package middleware

import (
	"bytes"
	"io"
	"time"

	"volcengine-go-server/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ResponseWriter 包装器，用于捕获响应内容
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logger 返回一个gin.HandlerFunc，用于记录HTTP请求日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 构建日志字段
		log := logger.GetLogger()
		fields := logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"status":     c.Writer.Status(),
			"duration":   duration.Milliseconds(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"size":       c.Writer.Size(),
		}

		// 如果有错误，添加错误信息
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// 根据状态码选择日志级别
		status := c.Writer.Status()
		if status >= 500 {
			log.WithFields(fields).Error("HTTP请求 - 服务器错误")
		} else if status >= 400 {
			log.WithFields(fields).Warn("HTTP请求 - 客户端错误")
		} else {
			log.WithFields(fields).Info("HTTP请求")
		}
	}
}

// DetailedLogger 返回一个详细的日志中间件，记录请求体和响应体
func DetailedLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，以便后续处理器可以读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应体捕获器
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)

		// 构建详细日志字段
		log := logger.GetLogger()
		fields := logrus.Fields{
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"status":        c.Writer.Status(),
			"duration":      duration.Milliseconds(),
			"ip":            c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"size":          c.Writer.Size(),
			"request_body":  string(requestBody),
			"response_body": responseWriter.body.String(),
		}

		// 如果有错误，添加错误信息
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// 根据状态码和响应时间选择日志级别
		status := c.Writer.Status()
		if status >= 500 {
			log.WithFields(fields).Error("HTTP请求详情 - 服务器错误")
		} else if status >= 400 {
			log.WithFields(fields).Warn("HTTP请求详情 - 客户端错误")
		} else if duration > 2*time.Second {
			log.WithFields(fields).Warn("HTTP请求详情 - 响应缓慢")
		} else {
			log.WithFields(fields).Info("HTTP请求详情")
		}
	}
}

// responseBodyWriter 用于捕获响应体
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Recovery 返回一个gin.HandlerFunc，用于从panic中恢复
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log := logger.GetLogger()
		log.WithFields(logrus.Fields{
			"error": recovered,
			"path":  c.Request.URL.Path,
		}).Error("服务器内部错误")
		c.AbortWithStatus(500)
	})
}
