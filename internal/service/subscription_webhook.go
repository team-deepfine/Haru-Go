package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/daewon/haru/pkg/appstore"
)

func (s *subscriptionService) HandleNotification(ctx context.Context, signedPayload string) error {
	notif, err := appstore.ParseSignedNotification(signedPayload)
	if err != nil {
		return fmt.Errorf("parse notification: %w", err)
	}

	slog.Info("apple notification received",
		"type", notif.NotificationType,
		"subtype", notif.Subtype,
		"originalTransactionId", notif.OriginalTransactionID,
	)

	user, err := s.userRepo.FindByOriginalTransactionID(ctx, notif.OriginalTransactionID)
	if err != nil {
		return fmt.Errorf("find user by original transaction id: %w", err)
	}

	switch notif.NotificationType {
	case appstore.NotificationDidRenew, appstore.NotificationSubscribed:
		user.SubscriptionStatus = "premium"
		if notif.ExpiresAt != nil {
			user.SubscriptionExpiry = notif.ExpiresAt
		}

	case appstore.NotificationExpired,
		appstore.NotificationRevoke,
		appstore.NotificationRefund,
		appstore.NotificationGracePeriodExpired:
		user.SubscriptionStatus = "free"

	case appstore.NotificationDidFailToRenew:
		// Billing retry 중 — 아직 premium 유지 (Apple grace period)
		slog.Info("billing retry in progress, keeping premium",
			"originalTransactionId", notif.OriginalTransactionID,
		)
		return nil

	default:
		slog.Info("unhandled notification type",
			"type", notif.NotificationType,
			"originalTransactionId", notif.OriginalTransactionID,
		)
		return nil
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update subscription status: %w", err)
	}

	slog.Info("subscription status updated",
		"originalTransactionId", notif.OriginalTransactionID,
		"status", user.SubscriptionStatus,
	)

	return nil
}
