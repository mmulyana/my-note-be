package users

import (
	"net/http"
	"os"
	"strings"

	"my-note-be/internal/constants"
	"my-note-be/internal/middleware"
	"my-note-be/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		service: NewService(db),
	}
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

	token, expiresAt, err := middleware.GenerateAccessToken(user.ID.String(), constants.AccessTokenTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken, err := h.service.CreateRefreshToken(user.ID, constants.RefreshTokenTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(c, "registered", TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Email:        user.Email,
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

	token, expiresAt, err := middleware.GenerateAccessToken(user.ID.String(), constants.AccessTokenTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken, err := h.service.CreateRefreshToken(user.ID, constants.RefreshTokenTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "logged in", TokenResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		Email:        user.Email,
	})
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var in RefreshTokenInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, newRefreshToken, err := h.service.RotateRefreshToken(in.RefreshToken, constants.RefreshTokenTTL)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	token, expiresAt, err := middleware.GenerateAccessToken(user.ID.String(), constants.AccessTokenTTL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "token refreshed", TokenResponse{
		AccessToken:  token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		Email:        user.Email,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	var in RefreshTokenInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.RevokeRefreshToken(in.RefreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "logged out", nil)
}

func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.service.FindByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}
	response.OK(c, "ok", ToProfileResponse(user))
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	var in UpdateProfileInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existing, err := h.service.FindByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	updated, err := h.service.UpdateProfile(userID, in)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if in.Photo != nil && existing.Photo != nil && *existing.Photo != "" && *existing.Photo != *in.Photo {
		os.Remove(strings.TrimPrefix(*existing.Photo, "/"))
	}

	response.OK(c, "updated", ToProfileResponse(updated))
}
