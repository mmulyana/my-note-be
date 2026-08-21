package folders

import (
	"errors"
	"net/http"
	"strconv"

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
	folders, err := h.service.FindAll(uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", ToResponses(folders))
}

func (h *Handler) FindAllWithNotes(c *gin.Context) {
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	page := 1
	if raw := c.Query("page"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			page = v
		}
	}

	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
			if limit > 100 {
				limit = 100
			}
		}
	}

	result, err := h.service.FindAllWithNotes(uid, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OKPaginated(c, "ok", ToWithNotesResponses(result.Folders, result.NotesByFolder, result.NoteCounts), page, limit, result.Total)
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
	f, err := h.service.FindOne(id, uid)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "ok", ToResponse(*f))
}

func (h *Handler) Create(c *gin.Context) {
	var in FolderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}
	f, err := h.service.Create(uid, in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.Created(c, "created", ToResponse(*f))
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
	var in FolderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	f, err := h.service.Update(id, uid, in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "updated", ToResponse(*f))
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
		response.Error(c, http.StatusNotFound, "folder not found")
	case errors.Is(err, ErrDuplicateName):
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}
