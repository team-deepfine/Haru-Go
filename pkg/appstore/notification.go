package appstore

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NotificationType represents Apple Server Notification V2 types.
type NotificationType string

const (
	NotificationDidRenew       NotificationType = "DID_RENEW"
	NotificationExpired        NotificationType = "EXPIRED"
	NotificationDidFailToRenew NotificationType = "DID_FAIL_TO_RENEW"
	NotificationRefund         NotificationType = "REFUND"
	NotificationRevoke         NotificationType = "REVOKE"
	NotificationSubscribed     NotificationType = "SUBSCRIBED"
	NotificationDidChangeRenewalPref NotificationType = "DID_CHANGE_RENEWAL_PREF"
	NotificationGracePeriodExpired   NotificationType = "GRACE_PERIOD_EXPIRED"
)

// NotificationSubtype represents the subtype of a notification.
type NotificationSubtype string

const (
	SubtypeAutoRenewEnabled  NotificationSubtype = "AUTO_RENEW_ENABLED"
	SubtypeAutoRenewDisabled NotificationSubtype = "AUTO_RENEW_DISABLED"
	SubtypeBillingRetryPeriod NotificationSubtype = "BILLING_RETRY_PERIOD"
	SubtypeVoluntary         NotificationSubtype = "VOLUNTARY"
	SubtypeBillingRecovery   NotificationSubtype = "BILLING_RECOVERY"
	SubtypeInitialBuy        NotificationSubtype = "INITIAL_BUY"
)

// ServerNotification represents the decoded Apple Server Notification V2 payload.
type ServerNotification struct {
	NotificationType    NotificationType    `json:"notificationType"`
	Subtype             NotificationSubtype `json:"subtype"`
	OriginalTransactionID string            `json:"originalTransactionId"`
	ExpiresAt           *time.Time          `json:"expiresAt,omitempty"`
	ProductID           string              `json:"productId"`
	IsRevoked           bool                `json:"isRevoked"`
}

// notificationPayload represents the outer signed payload from Apple.
type notificationPayload struct {
	NotificationType string `json:"notificationType"`
	Subtype          string `json:"subtype"`
	Data             struct {
		SignedTransactionInfo string `json:"signedTransactionInfo"`
	} `json:"data"`
}

// ParseSignedNotification verifies and parses an Apple Server Notification V2 signed payload.
func ParseSignedNotification(signedPayload string) (*ServerNotification, error) {
	// Parse the outer notification JWS
	payload, err := parseSignedJWS(signedPayload)
	if err != nil {
		return nil, fmt.Errorf("parse notification jws: %w", err)
	}

	var notif notificationPayload
	if err := json.Unmarshal(payload, &notif); err != nil {
		return nil, fmt.Errorf("unmarshal notification: %w", err)
	}

	if notif.Data.SignedTransactionInfo == "" {
		return nil, fmt.Errorf("empty signedTransactionInfo in notification")
	}

	// Parse the inner transaction JWS
	txPayload, err := parseSignedJWS(notif.Data.SignedTransactionInfo)
	if err != nil {
		return nil, fmt.Errorf("parse transaction jws: %w", err)
	}

	var txInfo TransactionInfo
	if err := json.Unmarshal(txPayload, &txInfo); err != nil {
		return nil, fmt.Errorf("unmarshal transaction info: %w", err)
	}

	result := &ServerNotification{
		NotificationType:      NotificationType(notif.NotificationType),
		Subtype:               NotificationSubtype(notif.Subtype),
		OriginalTransactionID: txInfo.OriginalTransactionID,
		ProductID:             txInfo.ProductID,
		IsRevoked:             txInfo.RevocationDate > 0,
	}

	if txInfo.ExpiresDate > 0 {
		t := time.UnixMilli(txInfo.ExpiresDate).UTC()
		result.ExpiresAt = &t
	}

	return result, nil
}

// parseSignedJWS verifies a JWS using the x5c certificate chain and returns the raw payload.
func parseSignedJWS(signed string) ([]byte, error) {
	token, err := jwt.Parse(signed, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		x5c, ok := token.Header["x5c"].([]interface{})
		if !ok || len(x5c) == 0 {
			return nil, fmt.Errorf("missing x5c header in JWS")
		}

		return verifyX5cChain(x5c)
	}, jwt.WithValidMethods([]string{"ES256"}))
	if err != nil {
		return nil, fmt.Errorf("verify jws signature: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return json.Marshal(claims)
}
