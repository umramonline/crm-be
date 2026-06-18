package config

import (
	"fmt"
	"time"

	env "github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                      int    `env:"PORT" envDefault:"8080"`
	CORSAllowedOrigins        string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:5173"`
	CORSAllowCredentials      bool   `env:"CORS_ALLOW_CREDENTIALS" envDefault:"true"`
	UmramonlineBaseURL        string `env:"UMRAMONLINE_BASE_URL"`
	UmramonlineAPIKey         string `env:"UMRAMONLINE_API_KEY"`
	UmramonlineOTPRequestPath string `env:"UMRAMONLINE_OTP_REQUEST_PATH" envDefault:"/api/v1/crm/auth/otp/request"`
	UmramonlineOTPVerifyPath  string `env:"UMRAMONLINE_OTP_VERIFY_PATH" envDefault:"/api/v1/crm/auth/otp/verify"`
	UmramonlinePasswordPath   string `env:"UMRAMONLINE_PASSWORD_LOGIN_PATH" envDefault:"/api/v1/crm/auth/password/login"`
	UmramonlineTimeoutSeconds int    `env:"UMRAMONLINE_TIMEOUT_SECONDS" envDefault:"10"`
	SessionTokenSecret        string `env:"SESSION_TOKEN_SECRET" envDefault:"dev-session-token-secret-change-me"`
	AccessTokenTTLMinutes     int    `env:"ACCESS_TOKEN_TTL_MINUTES" envDefault:"15"`
	RefreshTokenTTLDays       int    `env:"REFRESH_TOKEN_TTL_DAYS" envDefault:"30"`
	AuthCookieSecure          bool   `env:"AUTH_COOKIE_SECURE" envDefault:"false"`
	AuthCookieSameSite        string `env:"AUTH_COOKIE_SAME_SITE" envDefault:"Lax"`
	ShutdownTimeoutSeconds    int    `env:"SHUTDOWN_TIMEOUT_SECONDS" envDefault:"10"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c Config) UmramonlineTimeout() time.Duration {
	return time.Duration(c.UmramonlineTimeoutSeconds) * time.Second
}

func (c Config) AccessTokenTTL() time.Duration {
	return time.Duration(c.AccessTokenTTLMinutes) * time.Minute
}

func (c Config) RefreshTokenTTL() time.Duration {
	return time.Duration(c.RefreshTokenTTLDays) * 24 * time.Hour
}

func (c Config) ShutdownTimeout() time.Duration {
	return time.Duration(c.ShutdownTimeoutSeconds) * time.Second
}
