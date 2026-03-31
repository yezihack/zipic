package handler

import (
	"zipic/internal/version"
	"zipic/pkg/response"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all API routes with security middleware
func RegisterRoutes(r *gin.Engine, h *ImageHandler) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// Version info
	r.GET("/version", func(c *gin.Context) {
		response.Success(c, version.Info())
	})

	// API routes with rate limiting
	api := r.Group("/api")
	api.Use(RateLimitMiddleware())
	{
		// Upload route (single file)
		api.POST("/upload", FileSizeLimitMiddleware(), h.Upload)

		// Compress route (single file)
		api.POST("/compress", FileSizeLimitMiddleware(), h.Compress)

		// Batch compress route (multiple files)
		api.POST("/batch-compress", BatchSizeLimitMiddleware(), h.BatchCompress)

		// Download routes (no size limit needed)
		api.GET("/download", h.Download)
		api.GET("/download-zip", h.DownloadZip)
		api.GET("/preview", h.Preview)
	}
}