package notes

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

	var labelID *uuid.UUID
	if raw := c.Query("labelId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid labelId")
			return
		}
		labelID = &id
	}

	var folderID *uuid.UUID
	if raw := c.Query("folderId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid folderId")
			return
		}
		folderID = &id
	}

	// note: hasFolder=true => only notes inside a folder, false => only loose notes
	var hasFolder *bool
	if raw := c.Query("hasFolder"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid hasFolder")
			return
		}
		hasFolder = &v
	}

	var archived *bool
	if raw := c.Query("archived"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			archived = &v
		}
	}

	var pinned *bool
	if raw := c.Query("pinned"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			pinned = &v
		}
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

	search := c.Query("q")

	notes, total, err := h.service.FindAll(uid, labelID, folderID, hasFolder, archived, pinned, search, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OKPaginated(c, "ok", ToListItemResponses(notes), page, limit, total)
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
