package config

import "os"

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	Port                    string
	Env                     string
	DatabaseURL             string
	JWTSecret               string
	UploadsDir              string
	PublicBaseURL           string
	FirebaseCredsFile       string
	AnthropicAPIKey         string
	AnthropicOCRModel       string
	GeminiAPIKey            string
	GeminiOCRModel          string
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
	TokenEncryptionKey      string
}

func Load() Config {
	port := getEnv("PORT", "8080")
	return Config{
		Port:                    port,
		Env:                     getEnv("APP_ENV", "development"),
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5434/warrantykeeper?sslmode=disable"),
		JWTSecret:               getEnv("JWT_SECRET", "dev-secret-change-me"),
		UploadsDir:              getEnv("UPLOADS_DIR", "./data/uploads"),
		PublicBaseURL:           getEnv("PUBLIC_BASE_URL", "http://localhost:"+port),
		FirebaseCredsFile:       getEnv("FIREBASE_CREDENTIALS_FILE", ""),
		AnthropicAPIKey:         getEnv("ANTHROPIC_API_KEY", ""),
		AnthropicOCRModel:       getEnv("ANTHROPIC_OCR_MODEL", "claude-haiku-4-5-20251001"),
		GeminiAPIKey:            getEnv("GEMINI_API_KEY", ""),
		GeminiOCRModel:          getEnv("GEMINI_OCR_MODEL", "gemini-3.6-flash"),
		GoogleOAuthClientID:     getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
		GoogleOAuthClientSecret: getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		TokenEncryptionKey:      getEnv("TOKEN_ENCRYPTION_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
