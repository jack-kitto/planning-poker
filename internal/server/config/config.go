package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Server configuration
	Port          string `env:"PORT" envDefault:"8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	AppEnv        string `env:"APP_ENV" envDefault:"development"`
	SessionSecret string `env:"SESSION_SECRET,required"`

	// Database configuration
	DBHost     string `env:"BLUEPRINT_DB_HOST" envDefault:"localhost"`
	DBPort     int    `env:"BLUEPRINT_DB_PORT" envDefault:"5432"`
	DBDatabase string `env:"BLUEPRINT_DB_DATABASE" envDefault:"blueprint"`
	DBUsername string `env:"BLUEPRINT_DB_USERNAME" envDefault:"melkey"`
	DBPassword string `env:"BLUEPRINT_DB_PASSWORD" envDefault:"password1234"`
	DBSchema   string `env:"BLUEPRINT_DB_SCHEMA" envDefault:"public"`

	// OAuth configuration
	GoogleOAuthClientID     string `env:"GOOGLE_OAUTH_CLIENT_ID"`
	GoogleOAuthClientSecret string `env:"GOOGLE_OAUTH_CLIENT_SECRET"`
	GoogleOAuthCallbackUrl  string `env:"GOOGLE_OAUTH_CALLBACK_URL,required"`
	GitHubOAuthClientID     string `env:"GITHUB_OAUTH_CLIENT_ID"`
	GitHubOAuthClientSecret string `env:"GITHUB_OAUTH_CLIENT_SECRET"`
	GitHubOAuthCallbackUrl  string `env:"GITHUB_OAUTH_CALLBACK_URL,required"`

	// Email configuration
	EmailAPIURL         string `env:"EMAIL_API_URL" envDefault:"https://api.resend.com/emails"`
	EmailServerUser     string `env:"EMAIL_SERVER_USER" envDefault:"resend"`
	EmailServerPassword string `env:"EMAIL_SERVER_PASSWORD,required"`
	EmailServerHost     string `env:"EMAIL_SERVER_HOST" envDefault:"smtp.resend.com"`
	EmailServerPort     int    `env:"EMAIL_SERVER_PORT" envDefault:"587"`
	EmailFrom           string `env:"EMAIL_FROM,required"`

	// Session configuration
	SessionExpiration time.Duration `env:"SESSION_EXPIRATION" envDefault:"24h"`
	AUTH_BYPASS       string        `env:"AUTH_BYPASS" envDefault:"false"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

// DatabaseURL returns a formatted database connection string
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s",
		c.DBUsername,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBDatabase,
		c.DBSchema,
	)
}
