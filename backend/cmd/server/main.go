package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"zipic/internal/config"
	"zipic/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.Default()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", cfg.Server.Port)
	}

	// Set gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create router
	r := gin.Default()

	// Configure CORS
	r.Use(corsMiddleware())

	// Create handlers
	imageHandler := handler.NewImageHandler()

	// Start cleanup scheduler (runs every hour)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			handler.CleanupOldFiles("./uploads")
		}
	}()

	// Register routes
	handler.RegisterRoutes(r, imageHandler)

	// Serve static files for frontend (embedded)
	staticFS, err := WebdistFS()
	if err != nil {
		log.Fatalf("Failed to load embedded webdist: %v", err)
	}

	// Serve assets
	r.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		c.FileFromFS("assets"+filepath, http.FS(staticFS))
	})

	// Serve favicon
	r.GET("/favicon.svg", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "favicon.svg")
		if err != nil {
			c.Status(404)
			return
		}
		c.Data(200, "image/svg+xml", data)
	})

	// SPA fallback - serve index.html for all unmatched routes (except /api)
	r.NoRoute(func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.String(500, "Failed to read index.html")
			return
		}
		c.Data(200, "text/html; charset=utf-8", data)
	})

	// Start server
	addr := ":" + port
	fmt.Printf("Server starting on http://localhost%s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}