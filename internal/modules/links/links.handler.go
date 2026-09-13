package links

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	rows, total, err := h.service.FindAll(uid, strings.TrimSpace(c.Query("search")), page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OKPaginated(c, "ok", toResponses(rows), page, limit, total)
}

func (h *Handler) Preview(c *gin.Context) {
	if _, ok := helpers.ParseUserID(c); !ok {
		return
	}

	raw := strings.TrimSpace(c.Query("url"))
	if raw == "" {
		response.Error(c, http.StatusBadRequest, "url is required")
		return
	}

	preview, err := h.service.Preview(raw)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnsupportedScheme):
			response.Error(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrBlockedHost):
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			// 502, bukan 500 — yang gagal halaman remote-nya, bukan app-nya
			response.Error(c, http.StatusBadGateway, err.Error())
		}
		return
	}

	response.OK(c, "ok", preview)
}
