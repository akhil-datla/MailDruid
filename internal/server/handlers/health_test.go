package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type mockDB struct {
	err error
}

func (m *mockDB) Ping(_ context.Context) error {
	return m.err
}

func TestLiveness(t *testing.T) {
	e := echo.New()
	h := NewHealthHandler(&mockDB{}, "test-version")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Liveness(c); err != nil {
		t.Fatalf("Liveness error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body == "" {
		t.Error("expected non-empty body")
	}
}

func TestReadinessHealthy(t *testing.T) {
	e := echo.New()
	h := NewHealthHandler(&mockDB{err: nil}, "v1")

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Readiness(c); err != nil {
		t.Fatalf("Readiness error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestReadinessUnhealthy(t *testing.T) {
	e := echo.New()
	h := NewHealthHandler(&mockDB{err: errors.New("connection refused")}, "v1")

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Readiness(c); err != nil {
		t.Fatalf("Readiness error: %v", err)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}
