package handler

import (
	"net/http"

	"zipic/internal/handler"

	"github.com/gin-gonic/gin"
)

// Handler is the main entry point for Vercel serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Configure CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Create image handler
	imageHandler := handler.NewImageHandler()

	// API routes
	api := router.Group("/api")
	{
		api.POST("/upload", imageHandler.Upload)
		api.POST("/compress", imageHandler.Compress)
		api.POST("/batch-compress", imageHandler.BatchCompress)
		api.GET("/download", imageHandler.Download)
		api.GET("/download-zip", imageHandler.DownloadZip)
		api.GET("/preview", imageHandler.Preview)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Handle the request
	router.ServeHTTP(w, r)
}