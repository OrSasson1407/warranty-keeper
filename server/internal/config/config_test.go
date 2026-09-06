package config_test

import (
	"testing"

	"warrantykeeper/server/internal/config"
)

// clearAllEnv resets every variable config.Load reads to "" (which its
// getEnv treats the same as unset), so each test starts from a known state
// regardless of what's in the surrounding shell.
func clearAllEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"PORT", "APP_ENV", "DATABASE_URL", "JWT_SECRET",
		"UPLOADS_DIR", "PUBLIC_BASE_URL", "FIREBASE_CREDENTIALS_FILE",
		"ANTHROPIC_API_KEY", "ANTHROPIC_OCR_MODEL",
		"GEMINI_API_KEY", "GEMINI_OCR_MODEL",
	} {
		t.Setenv(key, "")
	}
}

func TestLoad_DefaultsWhenNoEnvVarsSet(t *testing.T) {
	clearAllEnv(t)

	cfg := config.Load()
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.JWTSecret != "dev-secret-change-me" {
		t.Errorf("JWTSecret = %q, want the documented dev default", cfg.JWTSecret)
	}
	if cfg.UploadsDir != "./data/uploads" {
		t.Errorf("UploadsDir = %q, want %q", cfg.UploadsDir, "./data/uploads")
	}
	if cfg.PublicBaseURL != "http://localhost:8080" {
		t.Errorf("PublicBaseURL = %q, want %q", cfg.PublicBaseURL, "http://localhost:8080")
	}
	if cfg.DatabaseURL == "" {
		t.Error("expected a non-empty default DatabaseURL")
	}
	if cfg.FirebaseCredsFile != "" {
		t.Errorf("FirebaseCredsFile = %q, want empty by default", cfg.FirebaseCredsFile)
	}
	if cfg.AnthropicAPIKey != "" {
		t.Errorf("AnthropicAPIKey = %q, want empty by default (no real OCR provider without one)", cfg.AnthropicAPIKey)
	}
	if cfg.AnthropicOCRModel != "claude-haiku-4-5-20251001" {
		t.Errorf("AnthropicOCRModel = %q, want the documented default", cfg.AnthropicOCRModel)
	}
	if cfg.GeminiAPIKey != "" {
		t.Errorf("GeminiAPIKey = %q, want empty by default", cfg.GeminiAPIKey)
	}
	if cfg.GeminiOCRModel != "gemini-2.0-flash" {
		t.Errorf("GeminiOCRModel = %q, want the documented default", cfg.GeminiOCRModel)
	}
}

func TestLoad_ReadsEnvVarsWhenSet(t *testing.T) {
	clearAllEnv(t)
	t.Setenv("PORT", "9999")
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://custom/db")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("UPLOADS_DIR", "/data/uploads")
	t.Setenv("PUBLIC_BASE_URL", "https://api.example.com")
	t.Setenv("FIREBASE_CREDENTIALS_FILE", "/etc/firebase.json")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test")
	t.Setenv("ANTHROPIC_OCR_MODEL", "claude-sonnet-5")
	t.Setenv("GEMINI_API_KEY", "gemini-test-key")
	t.Setenv("GEMINI_OCR_MODEL", "gemini-1.5-pro")

	cfg := config.Load()
	if cfg.Port != "9999" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9999")
	}
	if cfg.Env != "production" {
		t.Errorf("Env = %q, want %q", cfg.Env, "production")
	}
	if cfg.DatabaseURL != "postgres://custom/db" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://custom/db")
	}
	if cfg.JWTSecret != "super-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "super-secret")
	}
	if cfg.UploadsDir != "/data/uploads" {
		t.Errorf("UploadsDir = %q, want %q", cfg.UploadsDir, "/data/uploads")
	}
	if cfg.PublicBaseURL != "https://api.example.com" {
		t.Errorf("PublicBaseURL = %q, want %q", cfg.PublicBaseURL, "https://api.example.com")
	}
	if cfg.FirebaseCredsFile != "/etc/firebase.json" {
		t.Errorf("FirebaseCredsFile = %q, want %q", cfg.FirebaseCredsFile, "/etc/firebase.json")
	}
	if cfg.AnthropicAPIKey != "sk-ant-test" {
		t.Errorf("AnthropicAPIKey = %q, want %q", cfg.AnthropicAPIKey, "sk-ant-test")
	}
	if cfg.AnthropicOCRModel != "claude-sonnet-5" {
		t.Errorf("AnthropicOCRModel = %q, want %q", cfg.AnthropicOCRModel, "claude-sonnet-5")
	}
	if cfg.GeminiAPIKey != "gemini-test-key" {
		t.Errorf("GeminiAPIKey = %q, want %q", cfg.GeminiAPIKey, "gemini-test-key")
	}
	if cfg.GeminiOCRModel != "gemini-1.5-pro" {
		t.Errorf("GeminiOCRModel = %q, want %q", cfg.GeminiOCRModel, "gemini-1.5-pro")
	}
}

func TestLoad_PublicBaseURLDerivesFromCustomPortWhenNotSetExplicitly(t *testing.T) {
	clearAllEnv(t)
	t.Setenv("PORT", "9090")

	cfg := config.Load()
	if cfg.PublicBaseURL != "http://localhost:9090" {
		t.Errorf("PublicBaseURL = %q, want %q", cfg.PublicBaseURL, "http://localhost:9090")
	}
}

func TestLoad_ExplicitPublicBaseURLOverridesPortDerivation(t *testing.T) {
	clearAllEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("PUBLIC_BASE_URL", "https://warrantykeeper.example.com")

	cfg := config.Load()
	if cfg.PublicBaseURL != "https://warrantykeeper.example.com" {
		t.Errorf("PublicBaseURL = %q, want the explicit override, not one derived from PORT", cfg.PublicBaseURL)
	}
}
