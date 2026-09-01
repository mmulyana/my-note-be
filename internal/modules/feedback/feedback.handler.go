package feedback

import (
	"errors"
	"net/http"

	"my-note-be/internal/helpers"
	"my-note-be/internal/modules/users"
	"my-note-be/internal/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service     *Service
	userService *users.Service
}

func NewHandler(service *Service, userService *users.Service) *Handler {
	return &Handler{service: service, userService: userService}
}

func (h *Handler) Create(c *gin.Context) {
	var in FeedbackInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	uid, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	user, err := h.userService.FindByID(uid.String())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to load user")
		return
	}

	reporterName := user.Email
	if user.Username != nil && *user.Username != "" {
		reporterName = *user.Username
	}

	res, err := h.service.Create(in, reporterName, user.Email, user.ID.String())
	if err != nil {
		if errors.Is(err, ErrInvalidType) {
			response.Error(c, http.StatusBadRequest, "invalid feedback type")
			return
		}
		response.Error(c, http.StatusBadGateway, "failed to send feedback")
		return
	}

	response.Created(c, "created", res)
}
