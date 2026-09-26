package labels

import (
	"errors"
	"net/http"

	"my-note-be/internal/helpers"
	"my-note-be/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{service: NewService(db)}
}

func (h *Handler) FindAll(c *gin.Context) {
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	labels, err := h.service.FindAll(uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", ToResponses(labels))
}

func (h *Handler) FindOne(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	label, err := h.service.FindOne(id, uid)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "ok", ToResponse(*label))
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}
	return id, true
}

func respondLookupError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "label not found")
		return
	}
	response.Error(c, http.StatusInternalServerError, err.Error())
}
