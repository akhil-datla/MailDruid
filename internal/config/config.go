package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	AllowOrigins []string      `mapstructure:"allow_origins"`
	RateLimit    float64       `mapstructure:"rate_limit"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

type AuthConfig struct {
	SigningKey    string        `mapstructure:"signing_key"`
	EncryptionKey string        `mapstructure:"encryption_key"`
	TokenExpiry  time.Duration `mapstructure:"token_expiry"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads configuration from file, environment variables, and defaults.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.allow_origins", []string{"*"})
	v.SetDefault("server.rate_limit", 20)

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "password")
	v.SetDefault("database.name", "maildruid")
	v.SetDefault("database.sslmode", "disable")

	v.SetDefault("smtp.host", "")
	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.email", "")
	v.SetDefault("smtp.password", "")

	v.SetDefault("auth.signing_key", "")
	v.SetDefault("auth.encryption_key", "")
	v.SetDefault("auth.token_expiry", "24h")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")

	// Config file
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.maildruid")
		v.AddConfigPath("/etc/maildruid")
	}

	// Environment variables
	v.SetEnvPrefix("MAILDRUID")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Auth.SigningKey == "" {
		return fmt.Errorf("auth.signing_key is required (set MAILDRUID_AUTH_SIGNING_KEY)")
	}
	if c.Auth.EncryptionKey == "" {
		return fmt.Errorf("auth.encryption_key is required (set MAILDRUID_AUTH_ENCRYPTION_KEY)")
	}
	if len(c.Auth.EncryptionKey) != 16 && len(c.Auth.EncryptionKey) != 24 && len(c.Auth.EncryptionKey) != 32 {
		return fmt.Errorf("auth.encryption_key must be 16, 24, or 32 bytes for AES")
	}
	if c.SMTP.Email == "" {
		return fmt.Errorf("smtp.email is required (set MAILDRUID_SMTP_EMAIL)")
	}
	if c.SMTP.Password == "" {
		return fmt.Errorf("smtp.password is required (set MAILDRUID_SMTP_PASSWORD)")
	}
	if c.SMTP.Host == "" {
		return fmt.Errorf("smtp.host is required (set MAILDRUID_SMTP_HOST)")
	}
	return nil
}
