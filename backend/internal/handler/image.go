package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"zipic/pkg/response"

	"github.com/gin-gonic/gin"
)

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
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0700) // Only current user can access
	os.MkdirAll(filepath.Join(uploadDir, "compressed"), 0700)
	return &ImageHandler{uploadDir: uploadDir}
}

// UploadDir returns the upload directory path
func (h *ImageHandler) UploadDir() string {
	return h.uploadDir
}

// StartCleanupScheduler starts the periodic file cleanup task
func StartCleanupScheduler(uploadDir string) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		CleanupOldFiles(uploadDir)
	}
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
	Original      ImageInfo `json:"original"`
	Compressed    ImageInfo `json:"compressed"`
	CompressionRatio float64 `json:"compression_ratio"`
	Quality       int       `json:"quality"`
}

// Upload handles image upload
func (h *ImageHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		response.BadRequest(c, "Invalid file type. Only JPG, PNG, WebP are allowed")
		return
	}

	// Read content with size limit
	content, err := io.ReadAll(io.LimitReader(file, MaxFileSize+1))
	if err != nil {
		response.InternalError(c, "Failed to read file")
		return
	}

	// Check file size
	if len(content) > MaxFileSize {
		response.BadRequest(c, "File too large (max 20MB)")
		return
	}

	// Check image dimensions (prevent decompression bomb)
	config, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		response.BadRequest(c, "Invalid or unsupported image format")
		return
	}

	// Check total pixels
	totalPixels := config.Width * config.Height
	if totalPixels > MaxImagePixels {
		response.BadRequest(c, "Image dimensions too large (max 10000x10000)")
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFilename(header.Filename))
	originalPath := filepath.Join(h.uploadDir, filename)

	// Save original file with restricted permissions
	if err := os.WriteFile(originalPath, content, 0600); err != nil {
		response.InternalError(c, "Failed to save file")
		return
	}

	// Get image info
	info, err := getImageInfo(originalPath, filename)
	if err != nil {
		os.Remove(originalPath)
		response.InternalError(c, "Failed to read image info")
		return
	}

	response.Success(c, info)
}

// Compress handles image compression
func (h *ImageHandler) Compress(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded")
		return
	}
	defer file.Close()

	// Get quality parameter
	qualityStr := c.DefaultPostForm("quality", "80")
	quality, err := strconv.Atoi(qualityStr)
	if err != nil || quality < 1 || quality > 100 {
		quality = 80
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		response.BadRequest(c, "Invalid file type. Only JPG, PNG, WebP are allowed")
		return
	}

	// Read file content with size limit
	content, err := io.ReadAll(io.LimitReader(file, MaxFileSize+1))
	if err != nil {
		response.InternalError(c, "Failed to read file")
		return
	}

	// Check file size
	if len(content) > MaxFileSize {
		response.BadRequest(c, "File too large (max 20MB)")
		return
	}

	// Check image dimensions before decoding (prevent decompression bomb)
	config, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		response.BadRequest(c, "Invalid or unsupported image format")
		return
	}

	// Check total pixels
	totalPixels := config.Width * config.Height
	if totalPixels > MaxImagePixels {
		response.BadRequest(c, "Image dimensions too large (max 10000x10000)")
		return
	}

	// Decode image
	img, format, err := decodeImage(bytes.NewReader(content))
	if err != nil {
		response.BadRequest(c, "Invalid or unsupported image format")
		return
	}

	// Compress image
	compressed, err := compressImage(img, format, quality)
	if err != nil {
		response.InternalError(c, "Failed to compress image")
		return
	}

	// Generate filename
	timestamp := time.Now().UnixNano()
	originalFilename := fmt.Sprintf("%d_%s", timestamp, sanitizeFilename(header.Filename))
	compressedFilename := fmt.Sprintf("%d_compressed_%s", timestamp, sanitizeFilename(header.Filename))

	// Save original with restricted permissions
	originalPath := filepath.Join(h.uploadDir, originalFilename)
	if err := os.WriteFile(originalPath, content, 0600); err != nil {
		response.InternalError(c, "Failed to save original file")
		return
	}

	// Save compressed with restricted permissions
	compressedPath := filepath.Join(h.uploadDir, "compressed", compressedFilename)
	if err := os.WriteFile(compressedPath, compressed, 0600); err != nil {
		os.Remove(originalPath)
		response.InternalError(c, "Failed to save compressed file")
		return
	}

	// Build result
	originalInfo, _ := getImageInfo(originalPath, originalFilename)
	compressedInfo, _ := getImageInfo(compressedPath, compressedFilename)

	result := CompressedResult{
		Original:      originalInfo,
		Compressed:    compressedInfo,
		CompressionRatio: float64(len(compressed)) / float64(len(content)) * 100,
		Quality:       quality,
	}

	response.Success(c, result)
}

// BatchCompress handles batch image compression
func (h *ImageHandler) BatchCompress(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "Invalid form data")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.BadRequest(c, "No files uploaded")
		return
	}

	// Check file count limit
	if len(files) > MaxBatchFiles {
		response.BadRequest(c, fmt.Sprintf("Too many files (max %d allowed)", MaxBatchFiles))
		return
	}

	// Get quality parameter
	qualityStr := c.DefaultPostForm("quality", "80")
	quality, err := strconv.Atoi(qualityStr)
	if err != nil || quality < 1 || quality > 100 {
		quality = 80
	}

	var results []CompressedResult
	var compressedFiles []string

	for _, fileHeader := range files {
		// Validate file type
		contentType := fileHeader.Header.Get("Content-Type")
		if !isValidImageType(contentType) {
			continue
		}

		// Open file
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		// Read content with size limit
		content, err := io.ReadAll(io.LimitReader(file, MaxFileSize+1))
		file.Close()
		if err != nil {
			continue
		}

		// Check file size
		if len(content) > MaxFileSize {
			continue
		}

		// Check image dimensions before decoding (prevent decompression bomb)
		config, _, err := image.DecodeConfig(bytes.NewReader(content))
		if err != nil {
			continue
		}

		// Check total pixels
		totalPixels := config.Width * config.Height
		if totalPixels > MaxImagePixels {
			continue
		}

		// Decode image
		img, format, err := decodeImage(bytes.NewReader(content))
		if err != nil {
			continue
		}

		// Compress image
		compressed, err := compressImage(img, format, quality)
		if err != nil {
			continue
		}

		// Generate filename
		timestamp := time.Now().UnixNano()
		originalFilename := fmt.Sprintf("%d_%s", timestamp, sanitizeFilename(fileHeader.Filename))
		compressedFilename := fmt.Sprintf("%d_compressed_%s", timestamp, sanitizeFilename(fileHeader.Filename))

		// Save original with restricted permissions
		originalPath := filepath.Join(h.uploadDir, originalFilename)
		os.WriteFile(originalPath, content, 0600)

		// Save compressed with restricted permissions
		compressedPath := filepath.Join(h.uploadDir, "compressed", compressedFilename)
		os.WriteFile(compressedPath, compressed, 0600)

		compressedFiles = append(compressedFiles, compressedFilename)

		// Build result
		originalInfo, _ := getImageInfo(originalPath, originalFilename)
		compressedInfo, _ := getImageInfo(compressedPath, compressedFilename)

		result := CompressedResult{
			Original:      originalInfo,
			Compressed:    compressedInfo,
			CompressionRatio: float64(len(compressed)) / float64(len(content)) * 100,
			Quality:       quality,
		}

		results = append(results, result)

		// Release memory immediately
		content = nil
		compressed = nil
	}

	// Generate batch ID for ZIP download if more than 1 file
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

	response.Success(c, gin.H{
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
		response.BadRequest(c, "Filename is required")
		return
	}

	// Check if it's a compressed file
	var filePath string
	if strings.Contains(filename, "compressed_") {
		filePath = filepath.Join(h.uploadDir, "compressed", filepath.Base(filename))
	} else {
		filePath = filepath.Join(h.uploadDir, filepath.Base(filename))
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.BadRequest(c, "File not found")
		return
	}

	c.FileAttachment(filePath, filename)
}

// Preview serves image for preview
func (h *ImageHandler) Preview(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		response.BadRequest(c, "Filename is required")
		return
	}

	// Check if it's a compressed file
	var filePath string
	if strings.Contains(filename, "compressed_") {
		filePath = filepath.Join(h.uploadDir, "compressed", filepath.Base(filename))
	} else {
		filePath = filepath.Join(h.uploadDir, filepath.Base(filename))
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.Status(http.StatusNotFound)
		return
	}

	// Set proper Content-Type based on extension
	ext := strings.ToLower(filepath.Ext(filePath))
	contentType := "image/jpeg"
	if ext == ".png" {
		contentType = "image/png"
	} else if ext == ".webp" {
		contentType = "image/webp"
	}
	c.Header("Content-Type", contentType)
	c.Header("X-Content-Type-Options", "nosniff") // Prevent MIME sniffing

	c.File(filePath)
}

// DownloadZip handles batch download as ZIP
func (h *ImageHandler) DownloadZip(c *gin.Context) {
	batchID := c.Query("batch_id")
	if batchID == "" {
		response.BadRequest(c, "Batch ID is required")
		return
	}

	sessionMutex.RLock()
	session, exists := batchSessions[batchID]
	sessionMutex.RUnlock()

	if !exists {
		response.BadRequest(c, "Batch session not found or expired")
		return
	}

	// Create ZIP in memory
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, filename := range session.Files {
		filePath := filepath.Join(h.uploadDir, "compressed", filename)
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Clean filename for ZIP (remove timestamp prefix)
		cleanName := strings.TrimPrefix(filename, fmt.Sprintf("%d_compressed_", 0))
		if idx := strings.Index(cleanName, "_"); idx > 0 {
			cleanName = cleanName[idx+1:]
		}

		w, err := zipWriter.Create(cleanName)
		if err != nil {
			continue
		}
		w.Write(data)
	}

	zipWriter.Close()

	// Set response headers
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=compressed_images_%s.zip", batchID))
	c.Data(http.StatusOK, "application/zip", buf.Bytes())
}

// CleanupOldFiles removes files older than 1 day
func CleanupOldFiles(uploadDir string) {
	cutoff := time.Now().Add(-24 * time.Hour)

	// Cleanup uploads directory
	cleanupDir(uploadDir, cutoff)

	// Cleanup compressed directory
	cleanupDir(filepath.Join(uploadDir, "compressed"), cutoff)

	// Cleanup expired batch sessions
	sessionMutex.Lock()
	for id, session := range batchSessions {
		if time.Since(session.CreatedAt) > 24*time.Hour {
			delete(batchSessions, id)
		}
	}
	sessionMutex.Unlock()
}

func cleanupDir(dir string, cutoff time.Time) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
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

func saveFile(file multipart.File, path string) error {
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
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

func decodeImage(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}

func compressImage(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer

	// Note: We don't resize the image here to keep original dimensions
	// This allows proper before/after comparison in the frontend
	// File size reduction is achieved through JPEG quality compression

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	case "png":
		// For PNG, we convert to JPEG for better compression
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	case "webp":
		// Convert webp to jpeg for compression
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	default:
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func sanitizeFilename(filename string) string {
	// Remove potentially dangerous characters
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "..", "")
	return filename
}