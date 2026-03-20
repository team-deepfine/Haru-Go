package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/daewon/haru/internal/dto"
	"github.com/daewon/haru/internal/middleware"
	"github.com/daewon/haru/internal/model"
	"github.com/daewon/haru/internal/service"
	"github.com/daewon/haru/pkg/response"
	"github.com/gin-gonic/gin"
)

// VoiceHandler handles HTTP requests for voice parsing.
type VoiceHandler struct {
	svc     service.VoiceParsingService
	subsSvc service.SubscriptionService
}

// NewVoiceHandler creates a new VoiceHandler.
func NewVoiceHandler(svc service.VoiceParsingService, subsSvc service.SubscriptionService) *VoiceHandler {
	return &VoiceHandler{svc: svc, subsSvc: subsSvc}
}

// RegisterRoutes registers voice parsing routes on a Gin router group.
func (h *VoiceHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/events/parse-voice", h.ParseVoice)
}

// ParseVoice handles POST /api/events/parse-voice.
func (h *VoiceHandler) ParseVoice(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required")
		return
	}

	var req dto.ParseVoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.Text) == "" {
		response.Error(c, http.StatusBadRequest, model.ErrTextRequired.Error())
		return
	}

	// Check subscription / daily limit before parsing
	if err := h.subsSvc.CheckVoiceParseLimit(c.Request.Context(), userID); err != nil {
		if errors.Is(err, model.ErrVoiceParseLimit) {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		slog.Error("subscription check failed", "error", err)
		response.Error(c, http.StatusInternalServerError, "internal server error")
		return
	}

	result, err := h.svc.ParseVoice(c.Request.Context(), service.ParseVoiceInput{
		Text: req.Text,
	})
	if err != nil {
		handleVoiceServiceError(c, err)
		return
	}

	// Increment count only after successful parsing
	if err := h.subsSvc.IncrementVoiceParseCount(c.Request.Context(), userID); err != nil {
		slog.Error("failed to increment voice parse count", "error", err)
	}

	response.JSON(c, http.StatusOK, result)
}

func handleVoiceServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, model.ErrTextRequired):
		response.Error(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, model.ErrParsingFailed):
		response.Error(c, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, model.ErrAIServiceUnavailable):
		response.Error(c, http.StatusBadGateway, err.Error())
	default:
		slog.Error("internal error in voice parsing", "error", err)
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
