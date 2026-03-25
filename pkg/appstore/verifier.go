package appstore

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// VerifiedTransaction holds the result of a verified Apple transaction.
type VerifiedTransaction struct {
	TransactionID         string
	OriginalTransactionID string
	BundleID              string
	ProductID             string
	ExpiresAt             *time.Time
	IsRevoked             bool
}

// VerifyTransaction calls the App Store Server API and parses the signed transaction.
func (c *Client) VerifyTransaction(transactionID string) (*VerifiedTransaction, error) {
	body, err := c.GetTransactionInfo(transactionID)
	if err != nil {
		return nil, err
	}

	var resp struct {
		SignedTransactionInfo string `json:"signedTransactionInfo"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.SignedTransactionInfo == "" {
		return nil, fmt.Errorf("empty signedTransactionInfo")
	}

	txInfo, err := c.parseAndVerifyTransaction(resp.SignedTransactionInfo)
	if err != nil {
		return nil, fmt.Errorf("verify transaction: %w", err)
	}

	if txInfo.BundleID != c.bundleID {
		return nil, fmt.Errorf("bundle id mismatch: got %s, want %s", txInfo.BundleID, c.bundleID)
	}

	result := &VerifiedTransaction{
		TransactionID:         txInfo.TransactionID,
		OriginalTransactionID: txInfo.OriginalTransactionID,
		BundleID:              txInfo.BundleID,
		ProductID:             txInfo.ProductID,
		IsRevoked:             txInfo.RevocationDate > 0,
	}

	if txInfo.ExpiresDate > 0 {
		t := time.UnixMilli(txInfo.ExpiresDate).UTC()
		result.ExpiresAt = &t
	}

	return result, nil
}

// parseAndVerifyTransaction verifies the JWS signature using the x5c certificate chain
// embedded in the JWS header and extracts transaction info.
func (c *Client) parseAndVerifyTransaction(signedPayload string) (*TransactionInfo, error) {
	token, err := jwt.Parse(signedPayload, func(token *jwt.Token) (interface{}, error) {
		// App Store JWS uses ES256 with x5c certificate chain in the header
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
		return nil, fmt.Errorf("verify transaction signature: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("marshal claims: %w", err)
	}

	var info TransactionInfo
	if err := json.Unmarshal(claimsJSON, &info); err != nil {
		return nil, fmt.Errorf("unmarshal transaction info: %w", err)
	}

	return &info, nil
}
