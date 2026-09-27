package users

import (
	"errors"
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

	response.Created(c, "registered", ToTokenResponse(user, token, refreshToken, expiresAt))
}

func (h *Handler) Guest(c *gin.Context) {
	user, err := h.service.CreateGuest()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
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

	response.Created(c, "guest created", ToTokenResponse(user, token, refreshToken, expiresAt))
}

func (h *Handler) UpgradeGuest(c *gin.Context) {
	var in RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.service.UpgradeGuest(c.GetString("user_id"), in.Email, in.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotGuest):
			response.Error(c, http.StatusConflict, err.Error())
		case errors.Is(err, ErrEmailTaken):
			response.Error(c, http.StatusBadRequest, err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.OK(c, "upgraded", ToProfileResponse(user))
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

	response.OK(c, "logged in", ToTokenResponse(user, token, refreshToken, expiresAt))
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

	response.OK(c, "token refreshed", ToTokenResponse(user, token, newRefreshToken, expiresAt))
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

func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var in ChangePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.ChangePassword(userID, in.CurrentPassword, in.NewPassword); err != nil {
		switch {
		case errors.Is(err, ErrIsGuest):
			response.Error(c, http.StatusConflict, err.Error())
		case errors.Is(err, ErrCurrentPasswordWrong):
			response.Error(c, http.StatusBadRequest, err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.OK(c, "password changed", nil)
}
