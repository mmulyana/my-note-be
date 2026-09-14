package todos

import (
	"errors"
	"net/http"
	"strconv"
	"time"

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

	todos, total, err := h.service.FindAll(uid, labelID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OKPaginated(c, "ok", ToResponses(todos), page, limit, total)
}

func (h *Handler) FindGroupByCreatedAt(c *gin.Context) {
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	loc := time.UTC
	if raw := c.Query("tz"); raw != "" {
		l, err := time.LoadLocation(raw)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid tz")
			return
		}
		loc = l
	}

	from, ok := parseDateQuery(c, "from", loc)
	if !ok {
		return
	}
	to, ok := parseDateQuery(c, "to", loc)
	if !ok {
		return
	}
	if to.Before(from) {
		response.Error(c, http.StatusBadRequest, "to must not be before from")
		return
	}

	groups, err := h.service.FindGroupByCreatedAt(uid, from, to, loc)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", groups)
}

func parseDateQuery(c *gin.Context, key string, loc *time.Location) (time.Time, bool) {
	raw := c.Query(key)
	if raw == "" {
		response.Error(c, http.StatusBadRequest, key+" is required")
		return time.Time{}, false
	}
	d, err := time.ParseInLocation("2006-01-02", raw, loc)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid "+key)
		return time.Time{}, false
	}
	return d, true
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

func (h *Handler) FindGroupByToday(c *gin.Context) {
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	raw := c.Query("date")
	if raw == "" {
		response.Error(c, http.StatusBadRequest, "date is required")
		return
	}
	date, err := time.Parse("2006-01-02", raw)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid date")
		return
	}

	groups, err := h.service.FindGroupByToday(uid, date)
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
