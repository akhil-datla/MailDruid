package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// DBPinger is implemented by the database to check connectivity.
type DBPinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db      DBPinger
	version string
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db DBPinger, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

// Liveness returns 200 if the service is running.
// GET /healthz
func (h *HealthHandler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: h.version,
	})
}

// Readiness returns 200 if the service can serve requests (DB connected).
// GET /readyz
func (h *HealthHandler) Readiness(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  "database unreachable",
		})
	}

	return c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: h.version,
	})
}
