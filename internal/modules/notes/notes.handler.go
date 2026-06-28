package notes

import (
	"errors"
	"net/http"

	"my-note-be/internal/helpers"
	"my-note-be/internal/response"

	"github.com/gin-gonic/gin"
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

	notes, err := h.service.FindAll(uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", ToListItemResponses(notes))
}

func (h *Handler) FindOne(c *gin.Context) {
	id := c.Param("id")
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	note, err := h.service.FindOne(id, uid)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "ok", ToDetailResponse(*note))
}

func (h *Handler) Create(c *gin.Context) {
	var in CreateNoteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	note, err := h.service.Create(uid, in)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Created(c, "created", ToDetailResponse(*note))
}

func (h *Handler) Save(c *gin.Context) {
	id := c.Param("id")
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	var in SaveNoteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	note, err := h.service.Save(id, uid, in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "updated", ToDetailResponse(*note))
}

func (h *Handler) Remove(c *gin.Context) {
	id := c.Param("id")
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	if err := h.service.Remove(id, uid); err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "deleted", gin.H{"id": id})
}

func respondLookupError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, http.StatusNotFound, "note not found")
		return
	}
	response.Error(c, http.StatusInternalServerError, err.Error())
}
