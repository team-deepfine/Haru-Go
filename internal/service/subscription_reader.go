package service

import (
	"context"
	"fmt"

	"github.com/daewon/haru/internal/dto"
	"github.com/google/uuid"
)

func (s *subscriptionService) GetStatus(ctx context.Context, userID uuid.UUID) (*dto.SubscriptionResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	// Lazy check: expire subscription if needed
	if user.SubscriptionStatus == "premium" && !user.IsPremium() {
		user.SubscriptionStatus = "free"
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("update expired subscription: %w", err)
		}
	}

	return dto.ToSubscriptionResponse(user, s.voiceParseLimit), nil
}
