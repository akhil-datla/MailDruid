package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCustomValidatorRejectsInvalidRequest(t *testing.T) {
	e := echo.New()
	e.Validator = NewValidator()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"not-an-email"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var lr LoginRequest
	if err := c.Bind(&lr); err != nil {
		t.Fatalf("bind error: %v", err)
	}

	err := c.Validate(lr)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCustomValidatorAcceptsValidRequest(t *testing.T) {
	e := echo.New()
	e.Validator = NewValidator()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"test@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var lr LoginRequest
	if err := c.Bind(&lr); err != nil {
		t.Fatalf("bind error: %v", err)
	}

	err := c.Validate(lr)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestBindAndValidateRejectsBadJSON(t *testing.T) {
	e := echo.New()
	e.Validator = NewValidator()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var lr LoginRequest
	err := bindAndValidate(c, &lr)
	if err != nil {
		t.Fatalf("bindAndValidate should return nil (writes response itself), got: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
