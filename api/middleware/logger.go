package middleware

import (
	"bytes"
	"io"
	"time"

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

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用结构化日志记录HTTP请求
		fields := logrus.Fields{
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
			"status_code": param.StatusCode,
			"latency_ms":  param.Latency.Milliseconds(),
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"user_agent":  param.Request.UserAgent(),
			"body_size":   param.BodySize,
		}

		// 添加查询参数
		if param.Request.URL.RawQuery != "" {
			fields["query_params"] = param.Request.URL.RawQuery
		}

		// 添加错误信息（如果有）
		if param.ErrorMessage != "" {
			fields["error"] = param.ErrorMessage
		}

		// 根据状态码选择日志级别
		if param.StatusCode >= 500 {
			logrus.WithFields(fields).Error("HTTP请求 - 服务器错误")
		} else if param.StatusCode >= 400 {
			logrus.WithFields(fields).Warn("HTTP请求 - 客户端错误")
		} else {
			logrus.WithFields(fields).Info("HTTP请求")
		}

		return ""
	})
}

// DetailedLogger 详细的HTTP请求日志中间件，记录请求体和响应体
func DetailedLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装响应写入器
		responseWriter := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 构建详细日志字段
		fields := logrus.Fields{
			"timestamp":     start.Format(time.RFC3339),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status_code":   c.Writer.Status(),
			"latency_ms":    latency.Milliseconds(),
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"content_type":  c.Request.Header.Get("Content-Type"),
			"request_size":  len(requestBody),
			"response_size": responseWriter.body.Len(),
		}

		// 添加查询参数
		if c.Request.URL.RawQuery != "" {
			fields["query_params"] = c.Request.URL.RawQuery
		}

		// 添加请求头（选择性记录）
		if authHeader := c.Request.Header.Get("Authorization"); authHeader != "" {
			fields["has_auth"] = true
		}

		// 记录请求体（仅对POST/PUT/PATCH请求，且大小合理时）
		if len(requestBody) > 0 && len(requestBody) < 1024 &&
			(c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH") {
			fields["request_body"] = string(requestBody)
		}

		// 记录响应体（仅当大小合理时）
		if responseWriter.body.Len() > 0 && responseWriter.body.Len() < 1024 {
			fields["response_body"] = responseWriter.body.String()
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// 根据状态码和处理时间选择日志级别
		if c.Writer.Status() >= 500 {
			logrus.WithFields(fields).Error("HTTP请求详情 - 服务器错误")
		} else if c.Writer.Status() >= 400 {
			logrus.WithFields(fields).Warn("HTTP请求详情 - 客户端错误")
		} else if latency > 5*time.Second {
			logrus.WithFields(fields).Warn("HTTP请求详情 - 响应缓慢")
		} else {
			logrus.WithFields(fields).Info("HTTP请求详情")
		}
	}
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered interface{}) {
		logrus.WithFields(logrus.Fields{
			"error":     recovered,
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"client_ip": c.ClientIP(),
			"timestamp": time.Now().Format(time.RFC3339),
		}).Error("服务器内部错误 - Panic恢复")

		c.JSON(500, gin.H{
			"error":   "服务器内部错误",
			"message": "请稍后重试",
		})
	})
}
