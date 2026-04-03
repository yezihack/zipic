package handler

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed webdist
var webdist embed.FS

// ImageHandler handles image-related requests
type ImageHandler struct {
	uploadDir string
}

// BatchSession stores batch compression session data
type BatchSession struct {
	ID        string
	Files     []string
	CreatedAt time.Time
}

var (
	batchSessions = make(map[string]*BatchSession)
	sessionMutex  sync.RWMutex
)

// NewImageHandler creates a new ImageHandler
func NewImageHandler() *ImageHandler {
	uploadDir := filepath.Join(os.TempDir(), "zipic")
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(filepath.Join(uploadDir, "compressed"), 0755)
	return &ImageHandler{uploadDir: uploadDir}
}

// ImageInfo represents image information
type ImageInfo struct {
	Filename    string `json:"filename"`
	OriginalURL string `json:"original_url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Size        int64  `json:"size"`
	Format      string `json:"format"`
}

// CompressedResult represents compression result
type CompressedResult struct {
	Original         ImageInfo `json:"original"`
	Compressed       ImageInfo `json:"compressed"`
	CompressionRatio float64   `json:"compression_ratio"`
	Quality          int       `json:"quality"`
}

// Response represents unified API response
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Msg: "success", Data: data})
}

func badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{Code: 400, Msg: msg})
}

func internalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{Code: 500, Msg: msg})
}

// Compress handles image compression
func (h *ImageHandler) Compress(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		badRequest(c, "No file uploaded")
		return
	}
	defer file.Close()

	qualityStr := c.DefaultPostForm("quality", "80")
	quality, err := strconv.Atoi(qualityStr)
	if err != nil || quality < 1 || quality > 100 {
		quality = 80
	}

	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		badRequest(c, "Invalid file type. Only JPG, PNG, WebP are allowed")
		return
	}

	content, err := io.ReadAll(file)
	if err != nil {
		internalError(c, "Failed to read file")
		return
	}

	img, format, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		badRequest(c, "Failed to decode image: "+err.Error())
		return
	}

	compressed, err := compressImage(img, format, quality)
	if err != nil {
		internalError(c, "Failed to compress image: "+err.Error())
		return
	}

	timestamp := time.Now().UnixNano()
	originalFilename := fmt.Sprintf("%d_%s", timestamp, sanitizeFilename(header.Filename))
	compressedFilename := fmt.Sprintf("%d_compressed_%s", timestamp, sanitizeFilename(header.Filename))

	originalPath := filepath.Join(h.uploadDir, originalFilename)
	os.WriteFile(originalPath, content, 0644)

	compressedPath := filepath.Join(h.uploadDir, "compressed", compressedFilename)
	os.WriteFile(compressedPath, compressed, 0644)

	originalInfo, _ := getImageInfo(originalPath, originalFilename)
	compressedInfo, _ := getImageInfo(compressedPath, compressedFilename)

	result := CompressedResult{
		Original:         originalInfo,
		Compressed:       compressedInfo,
		CompressionRatio: float64(len(compressed)) / float64(len(content)) * 100,
		Quality:          quality,
	}

	success(c, result)
}

// BatchCompress handles batch image compression
func (h *ImageHandler) BatchCompress(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		badRequest(c, "Invalid form data")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		badRequest(c, "No files uploaded")
		return
	}

	qualityStr := c.DefaultPostForm("quality", "80")
	quality, err := strconv.Atoi(qualityStr)
	if err != nil || quality < 1 || quality > 100 {
		quality = 80
	}

	var results []CompressedResult
	var compressedFiles []string

	for _, fileHeader := range files {
		contentType := fileHeader.Header.Get("Content-Type")
		if !isValidImageType(contentType) {
			continue
		}

		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		content, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}

		img, format, err := image.Decode(bytes.NewReader(content))
		if err != nil {
			continue
		}

		compressed, err := compressImage(img, format, quality)
		if err != nil {
			continue
		}

		timestamp := time.Now().UnixNano()
		originalFilename := fmt.Sprintf("%d_%s", timestamp, sanitizeFilename(fileHeader.Filename))
		compressedFilename := fmt.Sprintf("%d_compressed_%s", timestamp, sanitizeFilename(fileHeader.Filename))

		originalPath := filepath.Join(h.uploadDir, originalFilename)
		os.WriteFile(originalPath, content, 0644)

		compressedPath := filepath.Join(h.uploadDir, "compressed", compressedFilename)
		os.WriteFile(compressedPath, compressed, 0644)

		compressedFiles = append(compressedFiles, compressedFilename)

		originalInfo, _ := getImageInfo(originalPath, originalFilename)
		compressedInfo, _ := getImageInfo(compressedPath, compressedFilename)

		result := CompressedResult{
			Original:         originalInfo,
			Compressed:       compressedInfo,
			CompressionRatio: float64(len(compressed)) / float64(len(content)) * 100,
			Quality:          quality,
		}

		results = append(results, result)
	}

	batchID := ""
	if len(compressedFiles) > 1 {
		batchID = fmt.Sprintf("%d", time.Now().UnixNano())
		sessionMutex.Lock()
		batchSessions[batchID] = &BatchSession{
			ID:        batchID,
			Files:     compressedFiles,
			CreatedAt: time.Now(),
		}
		sessionMutex.Unlock()
	}

	success(c, gin.H{
		"total":    len(files),
		"success":  len(results),
		"results":  results,
		"batch_id": batchID,
	})
}

// Download handles file download
func (h *ImageHandler) Download(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		badRequest(c, "Filename is required")
		return
	}

	var filePath string
	if strings.Contains(filename, "compressed_") {
		filePath = filepath.Join(h.uploadDir, "compressed", filepath.Base(filename))
	} else {
		filePath = filepath.Join(h.uploadDir, filepath.Base(filename))
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		badRequest(c, "File not found")
		return
	}

	c.FileAttachment(filePath, filename)
}

// DownloadZip handles batch download as ZIP
func (h *ImageHandler) DownloadZip(c *gin.Context) {
	batchID := c.Query("batch_id")
	if batchID == "" {
		badRequest(c, "Batch ID is required")
		return
	}

	sessionMutex.RLock()
	session, exists := batchSessions[batchID]
	sessionMutex.RUnlock()

	if !exists {
		badRequest(c, "Batch session not found or expired")
		return
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, filename := range session.Files {
		filePath := filepath.Join(h.uploadDir, "compressed", filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		cleanName := filename
		if idx := strings.Index(cleanName, "_compressed_"); idx > 0 {
			cleanName = cleanName[idx+12:]
		}

		w, err := zipWriter.Create(cleanName)
		if err != nil {
			continue
		}
		w.Write(data)
	}

	zipWriter.Close()

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=zipic_%s.zip", batchID))
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}

// Preview serves image for preview
func (h *ImageHandler) Preview(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		badRequest(c, "Filename is required")
		return
	}

	var filePath string
	if strings.Contains(filename, "compressed_") {
		filePath = filepath.Join(h.uploadDir, "compressed", filepath.Base(filename))
	} else {
		filePath = filepath.Join(h.uploadDir, filepath.Base(filename))
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.Status(http.StatusNotFound)
		return
	}

	c.File(filePath)
}

// Helper functions
func isValidImageType(contentType string) bool {
	validTypes := []string{"image/jpeg", "image/png", "image/webp"}
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	return false
}

func getImageInfo(path, filename string) (ImageInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return ImageInfo{}, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return ImageInfo{}, err
	}

	img, format, err := image.DecodeConfig(file)
	if err != nil {
		return ImageInfo{}, err
	}

	return ImageInfo{
		Filename:    filename,
		OriginalURL: "/api/preview?filename=" + filename,
		Width:       img.Width,
		Height:      img.Height,
		Size:        stat.Size(),
		Format:      format,
	}, nil
}

func compressImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	bounds := img.Bounds()
	maxDimension := 4096
	if bounds.Dx() > maxDimension || bounds.Dy() > maxDimension {
		if bounds.Dx() > bounds.Dy() {
			img = imaging.Resize(img, maxDimension, 0, imaging.Lanczos)
		} else {
			img = imaging.Resize(img, 0, maxDimension, imaging.Lanczos)
		}
	}

	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func sanitizeFilename(filename string) string {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "..", "")
	return filename
}

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	imageHandler := NewImageHandler()

	api := router.Group("/api")
	{
		api.POST("/compress", imageHandler.Compress)
		api.POST("/batch-compress", imageHandler.BatchCompress)
		api.GET("/download", imageHandler.Download)
		api.GET("/download-zip", imageHandler.DownloadZip)
		api.GET("/preview", imageHandler.Preview)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	staticFS, _ := fs.Sub(webdist, "webdist")

	router.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path[1:], http.FS(staticFS))
	})

	router.GET("/favicon.svg", func(c *gin.Context) {
		c.FileFromFS("favicon.svg", http.FS(staticFS))
	})

	router.NoRoute(func(c *gin.Context) {
		data, _ := fs.ReadFile(staticFS, "index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	router.ServeHTTP(w, r)
}