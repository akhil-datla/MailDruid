package config

import (
	"os"
	"testing"
)

func TestValidateRequiredFields(t *testing.T) {
	cfg := &Config{}
	err := cfg.validate()
	if err == nil {
		t.Error("expected validation error for empty config")
	}
}

func TestValidateEncryptionKeyLength(t *testing.T) {
	cfg := &Config{
		Auth: AuthConfig{
			SigningKey:    "test-signing-key",
			EncryptionKey: "invalid-length", // 14 bytes
		},
		SMTP: SMTPConfig{
			Email:    "test@test.com",
			Password: "pass",
			Host:     "smtp.test.com",
		},
	}
	err := cfg.validate()
	if err == nil {
		t.Error("expected error for invalid encryption key length")
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("MAILDRUID_AUTH_SIGNING_KEY", "test-key")
	os.Setenv("MAILDRUID_AUTH_ENCRYPTION_KEY", "0123456789abcdef")
	os.Setenv("MAILDRUID_SMTP_EMAIL", "test@test.com")
	os.Setenv("MAILDRUID_SMTP_PASSWORD", "pass")
	os.Setenv("MAILDRUID_SMTP_HOST", "smtp.test.com")
	os.Setenv("MAILDRUID_SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("MAILDRUID_AUTH_SIGNING_KEY")
		os.Unsetenv("MAILDRUID_AUTH_ENCRYPTION_KEY")
		os.Unsetenv("MAILDRUID_SMTP_EMAIL")
		os.Unsetenv("MAILDRUID_SMTP_PASSWORD")
		os.Unsetenv("MAILDRUID_SMTP_HOST")
		os.Unsetenv("MAILDRUID_SERVER_PORT")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Auth.SigningKey != "test-key" {
		t.Errorf("expected signing key 'test-key', got %q", cfg.Auth.SigningKey)
	}
}
