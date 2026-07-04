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

func (h *Handler) Create(c *gin.Context) {
	var in LabelInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	label, err := h.service.Create(uid, in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.Created(c, "created", ToResponse(*label))
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	var in LabelInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	label, err := h.service.Update(id, uid, in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "updated", ToResponse(*label))
}

func (h *Handler) Remove(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	if err := h.service.Remove(id, uid); err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "deleted", nil)
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
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.Error(c, http.StatusNotFound, "label not found")
	case errors.Is(err, ErrDuplicateName):
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}
