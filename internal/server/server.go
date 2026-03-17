package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/akhil-datla/maildruid/internal/config"
	"github.com/akhil-datla/maildruid/internal/domain/summary"
	"github.com/akhil-datla/maildruid/internal/domain/user"
	"github.com/akhil-datla/maildruid/internal/infrastructure/postgres"
	"github.com/akhil-datla/maildruid/internal/scheduler"
	"github.com/akhil-datla/maildruid/internal/server/handlers"
	"github.com/akhil-datla/maildruid/internal/server/middleware"
	"github.com/labstack/echo/v4"
	echoMW "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Server is the HTTP server.
type Server struct {
	echo   *echo.Echo
	cfg    config.ServerConfig
	logger *slog.Logger
}

// New creates and configures the HTTP server.
func New(
	cfg config.Config,
	db *postgres.DB,
	userSvc *user.Service,
	summarySvc *summary.Service,
	sched *scheduler.Scheduler,
	logger *slog.Logger,
) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Validator
	e.Validator = handlers.NewValidator()

	// Middleware
	e.Use(echoMW.Recover())
	e.Use(echoMW.RequestID())
	e.Use(echoMW.CORSWithConfig(echoMW.CORSConfig{
		AllowOrigins: cfg.Server.AllowOrigins,
		AllowMethods: []string{
			http.MethodGet, http.MethodPost,
			http.MethodPut, http.MethodPatch,
			http.MethodDelete, http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin, echo.HeaderContentType,
			echo.HeaderAccept, echo.HeaderAuthorization,
		},
	}))

	if cfg.Log.Level == "debug" {
		e.Use(echoMW.Logger())
	}

	// Rate limiter
	e.Use(echoMW.RateLimiter(echoMW.NewRateLimiterMemoryStore(
		rate.Limit(cfg.Server.RateLimit),
	)))

	// Handlers
	healthH := handlers.NewHealthHandler(db, Version)
	userH := handlers.NewUserHandler(userSvc, cfg.Auth)
	scheduleH := handlers.NewScheduleHandler(sched)
	summaryH := handlers.NewSummaryHandler(userSvc, summarySvc, logger)

	// Public routes
	e.GET("/healthz", healthH.Liveness)
	e.GET("/readyz", healthH.Readiness)

	// API v1
	v1 := e.Group("/api/v1")

	// Auth (public)
	v1.POST("/users", userH.Create)
	v1.POST("/auth/login", userH.Login)
	v1.GET("/schedules", scheduleH.List)

	// Protected routes
	auth := v1.Group("", middleware.JWTAuth([]byte(cfg.Auth.SigningKey)))

	// User management
	auth.GET("/users/me", userH.GetProfile)
	auth.PATCH("/users/me", userH.Update)
	auth.DELETE("/users/me", userH.Delete)

	// Email configuration
	auth.GET("/users/me/folders", userH.GetFolders)
	auth.PATCH("/users/me/folder", userH.UpdateFolder)
	auth.PUT("/users/me/tags", userH.UpdateTags)
	auth.PUT("/users/me/blacklist", userH.UpdateBlacklist)
	auth.PATCH("/users/me/start-time", userH.UpdateStartTime)
	auth.PATCH("/users/me/summary-count", userH.UpdateSummaryCount)

	// Scheduling
	auth.POST("/schedules", scheduleH.Create)
	auth.PATCH("/schedules", scheduleH.Update)
	auth.DELETE("/schedules", scheduleH.Delete)

	// Summary generation
	auth.POST("/summaries/generate", summaryH.Generate)

	// Serve embedded frontend (SPA fallback for non-API routes)
	serveFrontend(e)

	return &Server{echo: e, cfg: cfg.Server, logger: logger}
}

// Start begins serving HTTP requests.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.logger.Info("starting server", "address", addr, "version", Version)

	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}

	return s.echo.StartServer(srv)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	timeout := 10 * time.Second
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return s.echo.Shutdown(shutdownCtx)
}
