package dto

import (
	"time"

	"github.com/daewon/haru/internal/model"
)

// VerifySubscriptionRequest is the request body for subscription verification.
type VerifySubscriptionRequest struct {
	TransactionID string `json:"transactionId" binding:"required"`
}

// SubscriptionResponse is the response body for subscription status.
type SubscriptionResponse struct {
	SubscriptionStatus  string  `json:"subscriptionStatus"`
	ExpiresAt           *string `json:"expiresAt,omitempty"`
	VoiceParseCount     int     `json:"voiceParseCount"`
	VoiceParseLimit     int     `json:"voiceParseLimit"`
	VoiceParseRemaining int     `json:"voiceParseRemaining"`
}

// ToSubscriptionResponse converts a user model to a subscription response DTO.
func ToSubscriptionResponse(u *model.User, limit int) *SubscriptionResponse {
	resp := &SubscriptionResponse{
		SubscriptionStatus: u.SubscriptionStatus,
		VoiceParseLimit:    limit,
	}

	if u.SubscriptionExpiry != nil {
		s := u.SubscriptionExpiry.Format(time.RFC3339)
		resp.ExpiresAt = &s
	}

	if u.IsPremium() {
		resp.VoiceParseCount = 0
		resp.VoiceParseRemaining = limit
	} else {
		today := time.Now().In(time.FixedZone("KST", 9*60*60)).Truncate(24 * time.Hour)
		if u.VoiceParseDate != nil {
			parseDate := u.VoiceParseDate.In(time.FixedZone("KST", 9*60*60)).Truncate(24 * time.Hour)
			if parseDate.Equal(today) {
				resp.VoiceParseCount = u.VoiceParseCount
			}
		}
		remaining := limit - resp.VoiceParseCount
		if remaining < 0 {
			remaining = 0
		}
		resp.VoiceParseRemaining = remaining
	}

	return resp
}
