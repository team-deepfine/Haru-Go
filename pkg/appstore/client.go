package appstore

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/daewon/haru/pkg/applejwks"
	"github.com/golang-jwt/jwt/v5"
)

const (
	productionURL = "https://api.storekit.itunes.apple.com"
	sandboxURL    = "https://api.storekit-sandbox.itunes.apple.com"
)

// TransactionInfo holds the decoded transaction details from Apple.
type TransactionInfo struct {
	TransactionID       string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	BundleID            string `json:"bundleId"`
	ProductID           string `json:"productId"`
	ExpiresDate         int64  `json:"expiresDate"`
	RevocationDate      int64  `json:"revocationDate"`
}

// Client communicates with the App Store Server API v2.
type Client struct {
	privateKey   *ecdsa.PrivateKey
	keyID        string
	issuerID     string
	bundleID     string
	baseURL      string
	httpClient   *http.Client
	jwksVerifier *applejwks.Verifier
}

// NewClient creates a new App Store Server API client.
func NewClient(keyPath, keyID, issuerID, bundleID, environment string) (*Client, error) {
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from key file")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not ECDSA")
	}

	baseURL := sandboxURL
	if environment == "production" {
		baseURL = productionURL
	}

	jwksVerifier, _ := applejwks.NewVerifier()

	return &Client{
		privateKey:   ecKey,
		keyID:        keyID,
		issuerID:     issuerID,
		bundleID:     bundleID,
		baseURL:      baseURL,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		jwksVerifier: jwksVerifier,
	}, nil
}

// generateJWT creates a signed JWT for authenticating with the App Store Server API.
func (c *Client) generateJWT() (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": c.issuerID,
		"iat": now.Unix(),
		"exp": now.Add(20 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": c.bundleID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.keyID

	return token.SignedString(c.privateKey)
}

// GetTransactionInfo retrieves transaction info from the App Store Server API.
func (c *Client) GetTransactionInfo(transactionID string) ([]byte, error) {
	token, err := c.generateJWT()
	if err != nil {
		return nil, fmt.Errorf("generate apple jwt: %w", err)
	}

	url := fmt.Sprintf("%s/inApps/v1/transactions/%s", c.baseURL, transactionID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("app store api request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("app store api returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// BundleID returns the configured bundle ID for validation.
func (c *Client) BundleID() string {
	return c.bundleID
}
