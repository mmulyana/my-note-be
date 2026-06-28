package todos

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
	todos, err := h.service.FindAll(uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", ToResponses(todos))
}

func (h *Handler) FindGroupByNotes(c *gin.Context) {
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}
	groups, err := h.service.FindGroupByNotes(uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", groups)
}

func (h *Handler) FindGroupByDeadline(c *gin.Context) {
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}
	groups, err := h.service.FindGroupByDeadline(uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", groups)
}

func (h *Handler) FindOne(c *gin.Context) {
	id := c.Param("id")
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}
	t, err := h.service.FindOne(id, uid)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "ok", ToResponse(*t))
}

func (h *Handler) Create(c *gin.Context) {
	var in TodoInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}
	t, err := h.service.Create(uid, in)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, "note not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Created(c, "created", ToResponse(*t))
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}
	var in TodoUpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	t, err := h.service.Update(id, uid, in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "updated", ToResponse(*t))
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
		response.Error(c, http.StatusNotFound, "todo not found")
		return
	}
	response.Error(c, http.StatusInternalServerError, err.Error())
}
