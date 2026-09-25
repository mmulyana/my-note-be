package ai

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"my-note-be/internal/helpers"
	"my-note-be/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const streamTimeout = 2 * time.Minute

type Handler struct {
	service    *Service
	limiter    *limiter
	usage      *usageStore
	dailyLimit int64
}

func NewHandler(service *Service, db *gorm.DB, dailyLimit int) *Handler {
	if dailyLimit <= 0 {
		dailyLimit = defaultDailyTokens
	}
	return &Handler{
		service:    service,
		limiter:    newLimiter(limitRequests, limitWindow),
		usage:      &usageStore{db: db},
		dailyLimit: int64(dailyLimit),
	}
}

func (h *Handler) record(userID uuid.UUID, day string, messages []message, output string, reported int, canceled bool) {
	tokens := int64(reported)
	if tokens == 0 && (output != "" || canceled) {
		tokens = estimateTokens(messages, output)
	}
	if tokens == 0 {
		return
	}
	if err := h.usage.add(userID, day, tokens); err != nil {
		log.Printf("ai: record usage: %v", err)
	}
}

func (h *Handler) Usage(c *gin.Context) {
	userID, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	now := time.Now()
	day, resetsIn := usageDay(now)
	used, err := h.usage.used(userID, day)
	if err != nil {
		log.Printf("ai: read usage: %v", err)
		response.Error(c, http.StatusInternalServerError, "failed to read ai usage")
		return
	}

	response.OK(c, "ok", gin.H{
		"used":      used,
		"limit":     h.dailyLimit,
		"remaining": max(h.dailyLimit-used, 0),
		"resetsAt":  now.Add(resetsIn).UTC().Format(time.RFC3339),
	})
}

func (h *Handler) Stream(c *gin.Context) {
	userID, ok := helpers.ParseUserID(c)
	if !ok {
		return
	}

	var in StreamInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	messages, err := h.service.Build(in)
	if err != nil {
		if errors.Is(err, ErrNotConfigured) {
			response.Error(c, http.StatusServiceUnavailable, err.Error())
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	release, retryAfter, err := h.limiter.acquire(userID)
	if err != nil {
		c.Header("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
		response.Error(c, http.StatusTooManyRequests, err.Error())
		return
	}
	defer release()

	day, resetsIn := usageDay(time.Now())
	used, err := h.usage.used(userID, day)
	if err != nil {
		log.Printf("ai: read usage: %v", err)
		response.Error(c, http.StatusInternalServerError, "failed to check ai usage")
		return
	}
	if used >= h.dailyLimit {
		c.Header("Retry-After", strconv.Itoa(int(math.Ceil(resetsIn.Seconds()))))
		response.Error(c, http.StatusTooManyRequests, ErrQuotaExceeded.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), streamTimeout)
	defer cancel()

	started := false
	begin := func() {
		if started {
			return
		}
		started = true
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")
		c.Status(http.StatusOK)
		c.Writer.Flush()
	}

	var output strings.Builder
	reported, err := h.service.Stream(ctx, messages, func(delta string) error {
		output.WriteString(delta)
		begin()
		c.SSEvent("delta", gin.H{"text": delta})
		c.Writer.Flush()
		return ctx.Err()
	})

	canceled := errors.Is(err, context.Canceled)
	if err == nil || canceled || started {
		h.record(userID, day, messages, output.String(), reported, canceled)
	}

	if canceled {
		return
	}
	if err != nil {
		if !started {
			response.Error(c, http.StatusBadGateway, ErrUpstreamFailed.Error())
			return
		}
		c.SSEvent("error", gin.H{"message": ErrUpstreamFailed.Error()})
		c.Writer.Flush()
		return
	}

	begin()
	c.SSEvent("done", gin.H{})
	c.Writer.Flush()
}
