package conf

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Conf struct {
	Port         int    `env:"PORT" envDefault:"3000"`
	Password     string `env:"PASSWORD"`
	DatabasePath string `env:"DATABASE_PATH" envDefault:"data.db"`
	Environment  string `env:"ENVIRONMENT" envDefault:"development"`
	LogFormat    string `env:"LOG_FORMAT" envDefault:"text"` // "text" or "json"
	RateLimit    int    `env:"RATE_LIMIT" envDefault:"100"`  // requests per minute
}

func New() (*Conf, error) {
	cfg := Conf{}
	// Load .env file if it exists (optional for production)
	_ = godotenv.Load()

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Validate required fields
	if cfg.Password == "" {
		// In production, password should come from secrets manager
		if cfg.Environment == "production" {
			return nil, fmt.Errorf("PASSWORD environment variable is required in production")
		}
		// For development, allow empty password but warn
		fmt.Fprintf(os.Stderr, "WARNING: PASSWORD not set. This should be set in production.\n")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("PORT must be between 1 and 65535, got %d", cfg.Port)
	}

	return &cfg, nil
}
