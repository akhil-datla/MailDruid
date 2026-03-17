package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/akhil-datla/maildruid/internal/config"
	"github.com/akhil-datla/maildruid/internal/domain/user"
	"github.com/akhil-datla/maildruid/internal/infrastructure/encryption"
	"github.com/akhil-datla/maildruid/internal/server/handlers"
	"github.com/akhil-datla/maildruid/internal/server/middleware"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"

	echoMW "github.com/labstack/echo/v4/middleware"
)

// testEnv sets up a real Echo server with real services backed by an in-memory repo.
type testEnv struct {
	echo    *echo.Echo
	userSvc *user.Service
	authCfg config.AuthConfig
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	enc, err := encryption.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("encryption: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	repo := user.NewMemoryRepository()
	userSvc := user.NewService(repo, enc, logger)

	authCfg := config.AuthConfig{
		SigningKey:    "test-signing-key-32-bytes-long!!",
		EncryptionKey: "0123456789abcdef0123456789abcdef",
		TokenExpiry:   3600_000_000_000, // 1h in nanoseconds
	}

	e := echo.New()
	e.Validator = handlers.NewValidator()
	e.Use(echoMW.Recover())
	e.Use(echoMW.RateLimiter(echoMW.NewRateLimiterMemoryStore(rate.Limit(100))))

	userH := handlers.NewUserHandler(userSvc, authCfg)

	// Public routes
	v1 := e.Group("/api/v1")
	v1.POST("/users", userH.Create)
	v1.POST("/auth/login", userH.Login)

	// Protected routes
	auth := v1.Group("", middleware.JWTAuth([]byte(authCfg.SigningKey)))
	auth.GET("/users/me", userH.GetProfile)
	auth.PATCH("/users/me", userH.Update)
	auth.DELETE("/users/me", userH.Delete)
	auth.PATCH("/users/me/folder", userH.UpdateFolder)
	auth.PUT("/users/me/tags", userH.UpdateTags)
	auth.PUT("/users/me/blacklist", userH.UpdateBlacklist)
	auth.PATCH("/users/me/start-time", userH.UpdateStartTime)
	auth.PATCH("/users/me/summary-count", userH.UpdateSummaryCount)

	// Frontend
	e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("<!doctype html>"))
	})))

	return &testEnv{echo: e, userSvc: userSvc, authCfg: authCfg}
}

func (te *testEnv) request(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	te.echo.ServeHTTP(rec, req)
	return rec
}

func parseJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON response: %v\nbody: %s", err, rec.Body.String())
	}
	return result
}

// registerAndLogin is a helper that creates a user and returns a JWT token.
func registerAndLogin(t *testing.T, env *testEnv, email string) string {
	t.Helper()
	rec := env.request("POST", "/api/v1/users", map[string]interface{}{
		"name": "Test", "email": email, "receivingEmail": "r@t.com",
		"password": "secret123", "domain": "imap.t.com", "port": 993,
	}, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s: expected 201, got %d: %s", email, rec.Code, rec.Body.String())
	}

	rec = env.request("POST", "/api/v1/auth/login", map[string]interface{}{
		"email": email, "password": "secret123",
	}, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("login %s: expected 200, got %d: %s", email, rec.Code, rec.Body.String())
	}

	result := parseJSON(t, rec)
	token, ok := result["token"].(string)
	if !ok || token == "" {
		t.Fatalf("login %s: no token in response", email)
	}
	return token
}

// --- Tests ---

func TestFullUserLifecycle(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Register user
	rec := env.request("POST", "/api/v1/users", map[string]interface{}{
		"name":           "Jane Doe",
		"email":          "jane@test.com",
		"receivingEmail": "jane@gmail.com",
		"password":       "secret123",
		"domain":         "imap.test.com",
		"port":           993,
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Login
	rec = env.request("POST", "/api/v1/auth/login", map[string]interface{}{
		"email":    "jane@test.com",
		"password": "secret123",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	loginResult := parseJSON(t, rec)
	token, ok := loginResult["token"].(string)
	if !ok || token == "" {
		t.Fatal("login: expected non-empty token")
	}

	// 3. Get profile
	rec = env.request("GET", "/api/v1/users/me", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("get profile: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	profile := parseJSON(t, rec)
	if profile["name"] != "Jane Doe" {
		t.Errorf("profile name: expected 'Jane Doe', got %v", profile["name"])
	}
	if profile["email"] != "jane@test.com" {
		t.Errorf("profile email: expected 'jane@test.com', got %v", profile["email"])
	}
	if profile["receivingEmail"] != "jane@gmail.com" {
		t.Errorf("profile receivingEmail: expected 'jane@gmail.com', got %v", profile["receivingEmail"])
	}
	if profile["domain"] != "imap.test.com" {
		t.Errorf("profile domain: expected 'imap.test.com', got %v", profile["domain"])
	}
	// Password should NOT be in response (json:"-")
	if _, exists := profile["password"]; exists {
		t.Error("profile should not expose password")
	}

	// 4. Update profile
	rec = env.request("PATCH", "/api/v1/users/me", map[string]interface{}{
		"name":           "Jane Smith",
		"receivingEmail": "jane.smith@gmail.com",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update profile: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify update
	rec = env.request("GET", "/api/v1/users/me", nil, token)
	profile = parseJSON(t, rec)
	if profile["name"] != "Jane Smith" {
		t.Errorf("updated name: expected 'Jane Smith', got %v", profile["name"])
	}
	if profile["receivingEmail"] != "jane.smith@gmail.com" {
		t.Errorf("updated email: expected 'jane.smith@gmail.com', got %v", profile["receivingEmail"])
	}

	// 5. Update folder
	rec = env.request("PATCH", "/api/v1/users/me/folder", map[string]interface{}{
		"folder": "INBOX",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update folder: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 6. Update tags
	rec = env.request("PUT", "/api/v1/users/me/tags", map[string]interface{}{
		"tags": []string{"weekly", "report", "standup"},
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update tags: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 7. Update blacklist
	rec = env.request("PUT", "/api/v1/users/me/blacklist", map[string]interface{}{
		"senders": []string{"spam@co.com", "noreply@ads.com"},
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update blacklist: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 8. Update start time
	rec = env.request("PATCH", "/api/v1/users/me/start-time", map[string]interface{}{
		"startTime": "2025-01-01T00:00:00Z",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update start time: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 9. Update summary count
	rec = env.request("PATCH", "/api/v1/users/me/summary-count", map[string]interface{}{
		"count": 10,
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("update summary count: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 10. Verify all settings persisted
	rec = env.request("GET", "/api/v1/users/me", nil, token)
	profile = parseJSON(t, rec)
	if profile["folder"] != "INBOX" {
		t.Errorf("folder: expected 'INBOX', got %v", profile["folder"])
	}
	tags, ok := profile["tags"].([]interface{})
	if !ok || len(tags) != 3 {
		t.Errorf("tags: expected 3, got %v", profile["tags"])
	}
	bl, ok := profile["blackListSenders"].([]interface{})
	if !ok || len(bl) != 2 {
		t.Errorf("blacklist: expected 2, got %v", profile["blackListSenders"])
	}
	if profile["summaryCount"] != float64(10) {
		t.Errorf("summaryCount: expected 10, got %v", profile["summaryCount"])
	}

	// 11. Delete user
	rec = env.request("DELETE", "/api/v1/users/me", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 12. Verify user is gone
	rec = env.request("GET", "/api/v1/users/me", nil, token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("after delete: expected 404, got %d", rec.Code)
	}
}

func TestRegistrationValidation(t *testing.T) {
	env := setupTestEnv(t)

	tests := []struct {
		name string
		body map[string]interface{}
		code int
	}{
		{
			"missing name",
			map[string]interface{}{"email": "a@b.com", "receivingEmail": "a@b.com", "password": "123456", "domain": "imap.com", "port": 993},
			http.StatusBadRequest,
		},
		{
			"invalid email",
			map[string]interface{}{"name": "Test", "email": "not-email", "receivingEmail": "a@b.com", "password": "123456", "domain": "imap.com", "port": 993},
			http.StatusBadRequest,
		},
		{
			"password too short",
			map[string]interface{}{"name": "Test", "email": "a@b.com", "receivingEmail": "a@b.com", "password": "123", "domain": "imap.com", "port": 993},
			http.StatusBadRequest,
		},
		{
			"port out of range",
			map[string]interface{}{"name": "Test", "email": "a@b.com", "receivingEmail": "a@b.com", "password": "123456", "domain": "imap.com", "port": 99999},
			http.StatusBadRequest,
		},
		{
			"missing domain",
			map[string]interface{}{"name": "Test", "email": "a@b.com", "receivingEmail": "a@b.com", "password": "123456", "port": 993},
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := env.request("POST", "/api/v1/users", tt.body, "")
			if rec.Code != tt.code {
				t.Errorf("expected %d, got %d: %s", tt.code, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestLoginValidation(t *testing.T) {
	env := setupTestEnv(t)

	// Wrong credentials
	rec := env.request("POST", "/api/v1/auth/login", map[string]interface{}{
		"email":    "nobody@test.com",
		"password": "wrong",
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong credentials: expected 401, got %d", rec.Code)
	}

	// Missing fields
	rec = env.request("POST", "/api/v1/auth/login", map[string]interface{}{
		"email": "nobody@test.com",
	}, "")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("missing password: expected 400, got %d", rec.Code)
	}
}

func TestDuplicateRegistration(t *testing.T) {
	env := setupTestEnv(t)

	body := map[string]interface{}{
		"name": "User", "email": "dup@test.com", "receivingEmail": "r@t.com",
		"password": "secret123", "domain": "imap.test.com", "port": 993,
	}

	rec := env.request("POST", "/api/v1/users", body, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("first register: expected 201, got %d", rec.Code)
	}

	rec = env.request("POST", "/api/v1/users", body, "")
	if rec.Code != http.StatusConflict {
		t.Errorf("duplicate register: expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProtectedRoutesRequireAuth(t *testing.T) {
	env := setupTestEnv(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/users/me"},
		{"PATCH", "/api/v1/users/me"},
		{"DELETE", "/api/v1/users/me"},
		{"PATCH", "/api/v1/users/me/folder"},
		{"PUT", "/api/v1/users/me/tags"},
		{"PUT", "/api/v1/users/me/blacklist"},
		{"PATCH", "/api/v1/users/me/start-time"},
		{"PATCH", "/api/v1/users/me/summary-count"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			rec := env.request(ep.method, ep.path, nil, "")
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestInvalidTokenRejected(t *testing.T) {
	env := setupTestEnv(t)

	rec := env.request("GET", "/api/v1/users/me", nil, "totally-bogus-token")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid token, got %d", rec.Code)
	}
}

func TestTagsValidation(t *testing.T) {
	env := setupTestEnv(t)
	token := registerAndLogin(t, env, "tagval@t.com")

	// Empty tags array should fail validation
	rec := env.request("PUT", "/api/v1/users/me/tags", map[string]interface{}{
		"tags": []string{},
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("empty tags: expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSummaryCountValidation(t *testing.T) {
	env := setupTestEnv(t)
	token := registerAndLogin(t, env, "scval@t.com")

	// Zero count should fail
	rec := env.request("PATCH", "/api/v1/users/me/summary-count", map[string]interface{}{
		"count": 0,
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("zero count: expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	// Negative count should fail
	rec = env.request("PATCH", "/api/v1/users/me/summary-count", map[string]interface{}{
		"count": -5,
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("negative count: expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStartTimeValidation(t *testing.T) {
	env := setupTestEnv(t)
	token := registerAndLogin(t, env, "stval@t.com")

	// Invalid date format
	rec := env.request("PATCH", "/api/v1/users/me/start-time", map[string]interface{}{
		"startTime": "not-a-date",
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid date: expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestFrontendServing(t *testing.T) {
	env := setupTestEnv(t)

	// Non-API routes should serve frontend
	rec := env.request("GET", "/", nil, "")
	if rec.Code != http.StatusOK {
		t.Errorf("/ expected 200, got %d", rec.Code)
	}

	rec = env.request("GET", "/dashboard", nil, "")
	if rec.Code != http.StatusOK {
		t.Errorf("/dashboard expected 200, got %d", rec.Code)
	}

	rec = env.request("GET", "/settings", nil, "")
	if rec.Code != http.StatusOK {
		t.Errorf("/settings expected 200, got %d", rec.Code)
	}
}

func TestPasswordChangeFlow(t *testing.T) {
	env := setupTestEnv(t)

	// Register
	env.request("POST", "/api/v1/users", map[string]interface{}{
		"name": "Pass", "email": "pass@t.com", "receivingEmail": "r@t.com",
		"password": "oldpass123", "domain": "imap.t.com", "port": 993,
	}, "")

	// Login with old password
	rec := env.request("POST", "/api/v1/auth/login", map[string]interface{}{
		"email": "pass@t.com", "password": "oldpass123",
	}, "")
	token := parseJSON(t, rec)["token"].(string)

	// Change password
	rec = env.request("PATCH", "/api/v1/users/me", map[string]interface{}{
		"oldPassword": "oldpass123",
		"newPassword": "newpass456",
	}, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("password change: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Old password should fail
	rec = env.request("POST", "/api/v1/auth/login", map[string]interface{}{
		"email": "pass@t.com", "password": "oldpass123",
	}, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("old password: expected 401, got %d", rec.Code)
	}

	// New password should work
	rec = env.request("POST", "/api/v1/auth/login", map[string]interface{}{
		"email": "pass@t.com", "password": "newpass456",
	}, "")
	if rec.Code != http.StatusOK {
		t.Errorf("new password: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
