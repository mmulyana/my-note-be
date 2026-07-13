package uploads

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"my-note-be/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	baseUploadDir   = "uploads"
	defaultLocation = "misc"
	maxFileSize     = 5 * 1024 * 1024
)

var allowedExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

var validLocation = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "file is required")
		return
	}

	if file.Size > maxFileSize {
		response.Error(c, http.StatusBadRequest, "file must be smaller than 5MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		response.Error(c, http.StatusBadRequest, "file must be jpg, png, or webp")
		return
	}

	location := c.DefaultPostForm("location", defaultLocation)
	if !validLocation.MatchString(location) {
		response.Error(c, http.StatusBadRequest, "invalid location")
		return
	}

	dir := filepath.Join(baseUploadDir, location)
	if err := os.MkdirAll(dir, 0755); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to prepare upload directory")
		return
	}

	filename := uuid.New().String() + ext
	if err := c.SaveUploadedFile(file, filepath.Join(dir, filename)); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to save file")
		return
	}

	path := "/" + baseUploadDir + "/" + location + "/" + filename
	response.Created(c, "uploaded", UploadResponse{Path: path})
}
