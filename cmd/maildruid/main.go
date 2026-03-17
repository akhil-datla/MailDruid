package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/akhil-datla/maildruid/internal/config"
	"github.com/akhil-datla/maildruid/internal/domain/summary"
	"github.com/akhil-datla/maildruid/internal/domain/user"
	"github.com/akhil-datla/maildruid/internal/infrastructure/encryption"
	"github.com/akhil-datla/maildruid/internal/infrastructure/postgres"
	"github.com/akhil-datla/maildruid/internal/infrastructure/smtp"
	"github.com/akhil-datla/maildruid/internal/infrastructure/wordcloud"
	"github.com/akhil-datla/maildruid/internal/scheduler"
	"github.com/akhil-datla/maildruid/internal/server"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	cfgFile string
)

func main() {
	server.Version = version

	rootCmd := &cobra.Command{
		Use:   "maildruid",
		Short: "MailDruid - Automated Email Summary Service",
		Long: `MailDruid connects to your email via IMAP, summarizes messages
matching your configured tags, generates word cloud visualizations,
and delivers periodic summary digests to your inbox.`,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.yaml)")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MailDruid server",
		RunE:  runServe,
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE:  runMigrate,
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("MailDruid %s\n", version)
		},
	}

	rootCmd.AddCommand(serveCmd, migrateCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func setupLogger(cfg config.LogConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger := setupLogger(cfg.Log)
	logger.Info("starting MailDruid", "version", version)

	// Initialize encryption
	enc, err := encryption.New([]byte(cfg.Auth.EncryptionKey))
	if err != nil {
		return fmt.Errorf("initializing encryption: %w", err)
	}

	// Connect to database
	db, err := postgres.New(cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// Initialize services
	userRepo := postgres.NewUserRepository(db)
	userSvc := user.NewService(userRepo, enc, logger)

	// Locate font file relative to executable or CWD
	fontPath := findFontPath()
	generator := wordcloud.New(fontPath)
	summarySvc := summary.NewService(userSvc, generator, logger)

	mailer := smtp.New(cfg.SMTP)

	sched := scheduler.New(userSvc, summarySvc, mailer, logger)
	if err := sched.LoadExisting(cmd.Context()); err != nil {
		logger.Warn("failed to load existing tasks", "error", err)
	}

	// Create and start server
	srv := server.New(*cfg, db, userSvc, summarySvc, sched, logger)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		sched.Stop()
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error("server shutdown error", "error", err)
		}
		return nil
	case err := <-errCh:
		sched.Stop()
		return err
	}
}

func runMigrate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger := setupLogger(cfg.Log)

	db, err := postgres.New(cfg.Database, logger)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	logger.Info("migrations completed successfully")
	return nil
}

func findFontPath() string {
	// Check relative to executable
	exe, err := os.Executable()
	if err == nil {
		p := filepath.Join(filepath.Dir(exe), "fonts", "roboto", "Roboto-Regular.ttf")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Check CWD
	p := filepath.Join("fonts", "roboto", "Roboto-Regular.ttf")
	if _, err := os.Stat(p); err == nil {
		return p
	}

	return p
}
