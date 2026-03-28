package handler

import (
	"zipic/pkg/response"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.Engine, h *ImageHandler) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Image routes
		api.POST("/upload", h.Upload)
		api.POST("/compress", h.Compress)
		api.POST("/batch-compress", h.BatchCompress)
		api.GET("/download", h.Download)
		api.GET("/download-zip", h.DownloadZip)
		api.GET("/preview", h.Preview)
	}
}