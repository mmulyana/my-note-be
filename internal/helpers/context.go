package helpers

import (
	"net/http"

	"my-note-be/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ParseUserID(c *gin.Context) (uuid.UUID, bool) {
	uid, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return uuid.Nil, false
	}
	return uid, true
}
