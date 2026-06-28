package categories

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

	cats, err := h.service.FindAll(uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", ToResponses(cats))
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

	cat, err := h.service.FindOne(id, uid)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "ok", ToResponse(*cat))
}

func (h *Handler) Create(c *gin.Context) {
	var in CategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	cat, err := h.service.Create(uid, in)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Created(c, "created", ToResponse(*cat))
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

	var in CategoryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	cat, err := h.service.Update(id, uid, in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "updated", ToResponse(*cat))
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
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "category not found")
		return
	}
	response.Error(c, http.StatusInternalServerError, err.Error())
}
