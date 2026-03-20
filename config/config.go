package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// FCMConfig holds Firebase Cloud Messaging settings.
type FCMConfig struct {
	CredentialsFile string
	Enabled         bool
}

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port     string
	DB       DBConfig
	LogLevel string
	Gemini   GeminiConfig
	JWT      JWTConfig
	Apple        AppleConfig
	AppleIAP     AppleIAPConfig
	FCM          FCMConfig
	Subscription SubscriptionConfig
}

// DBConfig holds database connection settings.
type DBConfig struct {
	URL    string
	Driver string // "postgres" or "sqlite"
}

// GeminiConfig holds Gemini API settings.
type GeminiConfig struct {
	APIKey     string
	Model      string
	Timezone   string
	PromptPath string
}

// JWTConfig holds JWT signing and expiry settings.
type JWTConfig struct {
	Secret        string
	AccessExpiry  string
	RefreshExpiry string
}

// AppleConfig holds Apple Sign In OAuth settings.
type AppleConfig struct {
	ClientID    string
	TeamID      string
	KeyID       string
	PrivateKey  string
	RedirectURI string
}

// AppleIAPConfig holds Apple In-App Purchase verification settings.
type AppleIAPConfig struct {
	IssuerID       string
	KeyID          string
	PrivateKeyPath string
	BundleID       string
	Environment    string // "sandbox" or "production"
}

// SubscriptionConfig holds subscription settings.
type SubscriptionConfig struct {
	VoiceParseLimit int
}

// Load reads configuration from environment variables (with .env fallback).
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Port: getEnv("PORT", "8080"),
		DB: DBConfig{
			URL:    getEnv("DATABASE_URL", ""),
			Driver: getEnv("DB_DRIVER", "sqlite"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Gemini: GeminiConfig{
			APIKey:     getEnv("GEMINI_API_KEY", ""),
			Model:      getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
			Timezone:   getEnv("DEFAULT_TIMEZONE", "Asia/Seoul"),
			PromptPath: getEnv("VOICE_PARSE_PROMPT_PATH", ""),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", ""),
			AccessExpiry:  getEnv("JWT_ACCESS_EXPIRY", "1h"),
			RefreshExpiry: getEnv("JWT_REFRESH_EXPIRY", "720h"),
		},
		Apple: AppleConfig{
			ClientID:    getEnv("APPLE_CLIENT_ID", ""),
			TeamID:      getEnv("APPLE_TEAM_ID", ""),
			KeyID:       getEnv("APPLE_KEY_ID", ""),
			PrivateKey:  getEnv("APPLE_PRIVATE_KEY", ""),
			RedirectURI: getEnv("APPLE_REDIRECT_URI", ""),
		},
		AppleIAP: AppleIAPConfig{
			IssuerID:       getEnv("APPLE_IAP_ISSUER_ID", ""),
			KeyID:          getEnv("APPLE_IAP_KEY_ID", ""),
			PrivateKeyPath: getEnv("APPLE_IAP_PRIVATE_KEY_PATH", ""),
			BundleID:       getEnv("APPLE_IAP_BUNDLE_ID", ""),
			Environment:    getEnv("APPLE_IAP_ENVIRONMENT", "sandbox"),
		},
		FCM: FCMConfig{
			CredentialsFile: getEnv("FCM_CREDENTIALS_FILE", ""),
			Enabled:         getEnv("FCM_ENABLED", "false") == "true",
		},
		Subscription: SubscriptionConfig{
			VoiceParseLimit: getEnvInt("VOICE_PARSE_FREE_LIMIT", 3),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
