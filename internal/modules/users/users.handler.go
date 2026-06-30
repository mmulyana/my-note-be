package users

import (
	"net/http"

	"my-note-be/internal/middleware"
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

func (h *Handler) Register(c *gin.Context) {
	var in RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Register(in.Email, in.Password)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	token, expiresAt, err := middleware.GenerateToken(user.ID.String())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(c, "registered", TokenResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		Email:       user.Email,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var in LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.Login(in.Email, in.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	token, expiresAt, err := middleware.GenerateToken(user.ID.String())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "logged in", TokenResponse{
		AccessToken: token,
		ExpiresAt:   expiresAt,
		Email:       user.Email,
	})
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.service.FindByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}
	response.OK(c, "ok", ProfileResponse{
		ID:    user.ID.String(),
		Email: user.Email,
	})
}
