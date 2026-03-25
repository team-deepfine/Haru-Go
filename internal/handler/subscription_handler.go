package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/daewon/haru/internal/dto"
	"github.com/daewon/haru/internal/middleware"
	"github.com/daewon/haru/internal/model"
	"github.com/daewon/haru/internal/service"
	"github.com/daewon/haru/pkg/response"
	"github.com/gin-gonic/gin"
)

// SubscriptionHandler handles HTTP requests for subscription management.
type SubscriptionHandler struct {
	svc service.SubscriptionService
}

// NewSubscriptionHandler creates a new SubscriptionHandler.
func NewSubscriptionHandler(svc service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

// RegisterRoutes registers subscription routes on a Gin router group.
func (h *SubscriptionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/subscription/verify", h.Verify)
	rg.GET("/subscription", h.GetStatus)
}

// RegisterWebhookRoutes registers public webhook routes (no auth required).
func (h *SubscriptionHandler) RegisterWebhookRoutes(rg *gin.RouterGroup) {
	rg.POST("/apple/notifications", h.HandleAppleNotification)
}

// Verify handles POST /api/subscription/verify.
func (h *SubscriptionHandler) Verify(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required")
		return
	}

	var req dto.VerifySubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "transactionId is required")
		return
	}

	resp, err := h.svc.VerifyAndActivate(c.Request.Context(), userID, req.TransactionID)
	if err != nil {
		handleSubscriptionError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, resp)
}

// GetStatus handles GET /api/subscription.
func (h *SubscriptionHandler) GetStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "authentication required")
		return
	}

	resp, err := h.svc.GetStatus(c.Request.Context(), userID)
	if err != nil {
		handleSubscriptionError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, resp)
}

// HandleAppleNotification handles POST /api/apple/notifications.
func (h *SubscriptionHandler) HandleAppleNotification(c *gin.Context) {
	var req struct {
		SignedPayload string `json:"signedPayload" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "signedPayload is required")
		return
	}

	if err := h.svc.HandleNotification(c.Request.Context(), req.SignedPayload); err != nil {
		slog.Error("failed to handle apple notification", "error", err)
		// Always return 200 to Apple — non-200 causes Apple to retry indefinitely.
		// Errors are logged for investigation but should not trigger retries.
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "logged for investigation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleSubscriptionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, model.ErrUserNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, model.ErrInvalidTransaction):
		response.Error(c, http.StatusPaymentRequired, err.Error())
	case errors.Is(err, model.ErrStoreAPIFailed):
		response.Error(c, http.StatusBadGateway, err.Error())
	default:
		slog.Error("internal error in subscription", "error", err)
		response.Error(c, http.StatusInternalServerError, "internal server error")
	}
}
