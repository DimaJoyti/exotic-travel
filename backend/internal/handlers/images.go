package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ImageMetadata struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"originalName"`
	MimeType     string    `json:"mimeType"`
	Size         int64     `json:"size"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnailUrl,omitempty"`
	UploadedAt   time.Time `json:"uploadedAt"`
	UploadedBy   string    `json:"uploadedBy"`
	Tags         []string  `json:"tags,omitempty"`
	Alt          string    `json:"alt,omitempty"`
	Caption      string    `json:"caption,omitempty"`
}

type ImageUploadOptions struct {
	MaxSize           int64    `json:"maxSize"`
	AllowedTypes      []string `json:"allowedTypes"`
	Quality           float64  `json:"quality"`
	MaxWidth          int      `json:"maxWidth"`
	MaxHeight         int      `json:"maxHeight"`
	GenerateThumbnail bool     `json:"generateThumbnail"`
	ThumbnailSize     int      `json:"thumbnailSize"`
}

type ImageHandler struct {
	uploadDir string
	baseURL   string
}

func NewImageHandler(uploadDir, baseURL string) *ImageHandler {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}

	return &ImageHandler{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

// UploadImage handles image upload
func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse upload options
	var options ImageUploadOptions
	if optionsStr := r.FormValue("options"); optionsStr != "" {
		if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
			// Use default options if parsing fails
			options = getDefaultUploadOptions()
		}
	} else {
		options = getDefaultUploadOptions()
	}

	// Validate file
	if err := h.validateFile(header, options); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(h.uploadDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Create image metadata
	metadata := ImageMetadata{
		ID:           uuid.New().String(),
		Filename:     filename,
		OriginalName: header.Filename,
		MimeType:     header.Header.Get("Content-Type"),
		Size:         header.Size,
		Width:        1920, // Would be determined by image processing
		Height:       1080, // Would be determined by image processing
		URL:          fmt.Sprintf("%s/uploads/%s", h.baseURL, filename),
		UploadedAt:   time.Now(),
		UploadedBy:   "current_user", // Would come from auth context
		Tags:         []string{},
	}

	// Generate thumbnail if requested
	if options.GenerateThumbnail {
		thumbnailFilename := fmt.Sprintf("thumb_%s", filename)
		metadata.ThumbnailURL = fmt.Sprintf("%s/uploads/%s", h.baseURL, thumbnailFilename)
		// Thumbnail generation would happen here
	}

	// Return metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// GetImages returns paginated list of images
func (h *ImageHandler) GetImages(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	_ = r.URL.Query()["tags"]
	mimeType := r.URL.Query().Get("mimeType")
	_ = r.URL.Query().Get("uploadedBy")
	_ = r.URL.Query().Get("sortBy")
	_ = r.URL.Query().Get("sortOrder")

	// Mock data for now
	mockImages := []ImageMetadata{
		{
			ID:           "img_1",
			Filename:     "maldives_beach.jpg",
			OriginalName: "Beautiful Maldives Beach.jpg",
			MimeType:     "image/jpeg",
			Size:         2048576,
			Width:        1920,
			Height:       1080,
			URL:          fmt.Sprintf("%s/uploads/maldives_beach.jpg", h.baseURL),
			ThumbnailURL: fmt.Sprintf("%s/uploads/thumb_maldives_beach.jpg", h.baseURL),
			UploadedAt:   time.Now().Add(-24 * time.Hour),
			UploadedBy:   "admin",
			Tags:         []string{"beach", "maldives", "tropical"},
			Alt:          "Beautiful beach in Maldives with crystal clear water",
			Caption:      "Paradise found in the Maldives",
		},
		{
			ID:           "img_2",
			Filename:     "amazon_rainforest.jpg",
			OriginalName: "Amazon Rainforest Canopy.jpg",
			MimeType:     "image/jpeg",
			Size:         3145728,
			Width:        2048,
			Height:       1365,
			URL:          fmt.Sprintf("%s/uploads/amazon_rainforest.jpg", h.baseURL),
			ThumbnailURL: fmt.Sprintf("%s/uploads/thumb_amazon_rainforest.jpg", h.baseURL),
			UploadedAt:   time.Now().Add(-48 * time.Hour),
			UploadedBy:   "admin",
			Tags:         []string{"rainforest", "amazon", "nature", "green"},
			Alt:          "Lush Amazon rainforest canopy",
			Caption:      "The heart of the Amazon rainforest",
		},
	}

	// Apply filters (simplified for demo)
	filteredImages := mockImages
	if mimeType != "" {
		var filtered []ImageMetadata
		for _, img := range filteredImages {
			if img.MimeType == mimeType {
				filtered = append(filtered, img)
			}
		}
		filteredImages = filtered
	}

	// Calculate pagination
	total := len(filteredImages)
	start := (page - 1) * limit
	end := start + limit
	if end > total {
		end = total
	}
	if start > total {
		start = total
	}

	paginatedImages := filteredImages[start:end]

	response := map[string]interface{}{
		"images": paginatedImages,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetImage returns a single image by ID
func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	imageID := r.PathValue("id")

	// Mock data for now
	mockImage := ImageMetadata{
		ID:           imageID,
		Filename:     "sample_image.jpg",
		OriginalName: "Sample Image.jpg",
		MimeType:     "image/jpeg",
		Size:         2048576,
		Width:        1920,
		Height:       1080,
		URL:          fmt.Sprintf("%s/uploads/sample_image.jpg", h.baseURL),
		ThumbnailURL: fmt.Sprintf("%s/uploads/thumb_sample_image.jpg", h.baseURL),
		UploadedAt:   time.Now(),
		UploadedBy:   "admin",
		Tags:         []string{"sample", "demo"},
		Alt:          "Sample image for demonstration",
		Caption:      "This is a sample image",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mockImage)
}

// UpdateImageMetadata updates image metadata
func (h *ImageHandler) UpdateImageMetadata(w http.ResponseWriter, r *http.Request) {
	imageID := r.PathValue("id")

	var updateData struct {
		Alt     string   `json:"alt"`
		Caption string   `json:"caption"`
		Tags    []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Mock response
	updatedImage := ImageMetadata{
		ID:           imageID,
		Filename:     "sample_image.jpg",
		OriginalName: "Sample Image.jpg",
		MimeType:     "image/jpeg",
		Size:         2048576,
		Width:        1920,
		Height:       1080,
		URL:          fmt.Sprintf("%s/uploads/sample_image.jpg", h.baseURL),
		ThumbnailURL: fmt.Sprintf("%s/uploads/thumb_sample_image.jpg", h.baseURL),
		UploadedAt:   time.Now(),
		UploadedBy:   "admin",
		Tags:         updateData.Tags,
		Alt:          updateData.Alt,
		Caption:      updateData.Caption,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedImage)
}

// DeleteImage deletes an image
func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	imageID := r.PathValue("id")

	// In a real implementation, you would:
	// 1. Find the image record in the database
	// 2. Delete the physical file(s)
	// 3. Delete the database record

	fmt.Printf("Deleting image: %s\n", imageID)

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func getDefaultUploadOptions() ImageUploadOptions {
	return ImageUploadOptions{
		MaxSize:           10 * 1024 * 1024, // 10MB
		AllowedTypes:      []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
		Quality:           0.8,
		MaxWidth:          2048,
		MaxHeight:         2048,
		GenerateThumbnail: true,
		ThumbnailSize:     300,
	}
}

func (h *ImageHandler) validateFile(header *multipart.FileHeader, options ImageUploadOptions) error {
	// Check file size
	if header.Size > options.MaxSize {
		return fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", header.Size, options.MaxSize)
	}

	// Check file type
	mimeType := header.Header.Get("Content-Type")
	allowed := false
	for _, allowedType := range options.AllowedTypes {
		if mimeType == allowedType {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("file type %s is not allowed. Allowed types: %s", mimeType, strings.Join(options.AllowedTypes, ", "))
	}

	return nil
}
