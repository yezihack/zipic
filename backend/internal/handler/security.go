package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Security constants
const (
	MaxFileSize    = 20 * 1024 * 1024 // 20MB per file
	MaxBatchFiles  = 10               // max 10 files per batch
	MaxTotalSize   = MaxFileSize * MaxBatchFiles + 10*1024 // total request size for batch
	MaxImagePixels = 100_000_000      // 100 million pixels (max ~10000x10000)
)

// Global rate limiter: 10 requests per second, burst up to 20
var limiter = rate.NewLimiter(10, 20)

// RateLimitMiddleware limits request rate to prevent DoS attacks
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": 429,
				"msg":  "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// BodySizeLimitMiddleware limits request body size
func BodySizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// FileSizeLimitMiddleware for single file upload
func FileSizeLimitMiddleware() gin.HandlerFunc {
	return BodySizeLimitMiddleware(MaxFileSize + 1024) // extra for form fields
}

// BatchSizeLimitMiddleware for batch upload
func BatchSizeLimitMiddleware() gin.HandlerFunc {
	return BodySizeLimitMiddleware(MaxTotalSize)
}