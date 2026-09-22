package releases

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"my-note-be/internal/helpers"
	"my-note-be/internal/modules/users"
	"my-note-be/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	service     *Service
	userService *users.Service
	adminEmails map[string]struct{}
}

func NewHandler(db *gorm.DB, adminEmails []string) *Handler {
	set := make(map[string]struct{}, len(adminEmails))
	for _, email := range adminEmails {
		email = strings.ToLower(strings.TrimSpace(email))
		if email != "" {
			set[email] = struct{}{}
		}
	}
	return &Handler{
		service:     NewService(db),
		userService: users.NewService(db),
		adminEmails: set,
	}
}

func (h *Handler) FindAll(c *gin.Context) {
	if _, ok := helpers.ParseUserID(c); !ok {
		return
	}

	page := 1
	if raw := c.Query("page"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			page = v
		}
	}

	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
			if limit > 100 {
				limit = 100
			}
		}
	}

	list, total, err := h.service.FindAll(page, limit, h.isAdmin(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OKPaginated(c, "ok", ToResponses(list), page, limit, total)
}

func (h *Handler) FindOne(c *gin.Context) {
	if _, ok := helpers.ParseUserID(c); !ok {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	r, err := h.service.FindOne(id, h.isAdmin(c))
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "ok", ToResponse(*r))
}

func (h *Handler) Unread(c *gin.Context) {
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	count, seenAt, err := h.service.Unread(uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "ok", UnreadResponse{Count: count, LastSeenAt: seenAt})
}

func (h *Handler) MarkSeen(c *gin.Context) {
	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	seenAt, err := h.service.MarkSeen(uid)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "ok", UnreadResponse{Count: 0, LastSeenAt: &seenAt})
}

func (h *Handler) Create(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	var in Input
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	r, err := h.service.Create(in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.Created(c, "created", ToResponse(*r))
}

func (h *Handler) Update(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in Input
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	r, err := h.service.Update(id, in)
	if err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "updated", ToResponse(*r))
}

func (h *Handler) Remove(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.service.Remove(id); err != nil {
		respondLookupError(c, err)
		return
	}
	response.OK(c, "deleted", nil)
}

func (h *Handler) isAdmin(c *gin.Context) bool {
	if len(h.adminEmails) == 0 {
		return false
	}
	user, err := h.userService.FindByID(c.GetString("user_id"))
	if err != nil {
		return false
	}
	_, ok := h.adminEmails[strings.ToLower(user.Email)]
	return ok
}

func (h *Handler) requireAdmin(c *gin.Context) bool {
	if _, ok := helpers.ParseUserID(c); !ok {
		return false
	}
	if !h.isAdmin(c) {
		response.Error(c, http.StatusForbidden, "not allowed to edit releases")
		return false
	}
	return true
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
		response.Error(c, http.StatusNotFound, "release not found")
	case errors.Is(err, ErrInvalidTag):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}
