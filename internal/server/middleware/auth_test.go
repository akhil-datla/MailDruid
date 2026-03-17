package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestGenerateAndValidateToken(t *testing.T) {
	key := []byte("test-signing-key-for-jwt")

	token, err := GenerateToken("user-123", key, 1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Validate via middleware
	e := echo.New()
	handler := JWTAuth(key)(func(c echo.Context) error {
		id := GetUserID(c)
		if id != "user-123" {
			t.Errorf("expected user_id 'user-123', got %q", id)
		}
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler(c); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestJWTAuthRejectsMissingHeader(t *testing.T) {
	key := []byte("test-key")
	e := echo.New()

	handler := JWTAuth(key)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for missing auth header")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestJWTAuthRejectsInvalidToken(t *testing.T) {
	key := []byte("test-key")
	e := echo.New()

	handler := JWTAuth(key)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", he.Code)
	}
}

func TestJWTAuthRejectsExpiredToken(t *testing.T) {
	key := []byte("test-key")

	// Generate a token that expires immediately
	token, err := GenerateToken("user-456", key, -1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	e := echo.New()
	handler := JWTAuth(key)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler(c)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTAuthRejectsBadFormat(t *testing.T) {
	key := []byte("test-key")
	e := echo.New()

	handler := JWTAuth(key)(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "NotBearer some-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	if err == nil {
		t.Fatal("expected error for bad auth format")
	}
}
