package config

import (
	"fmt"
	"time"

	env "github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                               int    `env:"PORT" envDefault:"8080"`
	CORSAllowedOrigins                 string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:5173"`
	CORSAllowCredentials               bool   `env:"CORS_ALLOW_CREDENTIALS" envDefault:"true"`
	UmramonlineBaseURL                 string `env:"UMRAMONLINE_BASE_URL"`
	UmramonlineAPIKey                  string `env:"UMRAMONLINE_API_KEY"`
	UmramonlineAPIToken                string `env:"UMRAMONLINE_API_TOKEN"`
	UmramonlineOTPRequestPath          string `env:"UMRAMONLINE_OTP_REQUEST_PATH" envDefault:"/api/v1/crm/auth/otp/request"`
	UmramonlineOTPVerifyPath           string `env:"UMRAMONLINE_OTP_VERIFY_PATH" envDefault:"/api/v1/crm/auth/otp/verify"`
	UmramonlinePasswordPath            string `env:"UMRAMONLINE_PASSWORD_LOGIN_PATH" envDefault:"/api/v1/crm/auth/password/login"`
	UmramonlineUserRolesPath           string `env:"UMRAMONLINE_USER_ROLES_PATH" envDefault:"/api/v1/crm/auth/user-roles"`
	UmramonlineCustomersPath           string `env:"UMRAMONLINE_CUSTOMERS_PATH" envDefault:"/api/v1/crm/customers"`
	UmramonlineCustomerSearchPath      string `env:"UMRAMONLINE_CUSTOMER_SEARCH_PATH" envDefault:"/api/v1/crm/customers/search"`
	UmramonlineCustomerPhoneExistsPath string `env:"UMRAMONLINE_CUSTOMER_PHONE_EXISTS_PATH" envDefault:"/api/v1/crm/customers/phone-exists"`
	UmramonlineZonesPath               string `env:"UMRAMONLINE_ZONES_PATH" envDefault:"/api/v1/crm/zones"`
	UmramonlineCitiesPath              string `env:"UMRAMONLINE_CITIES_PATH" envDefault:"/api/v1/crm/cities"`
	UmramonlineTownsPath               string `env:"UMRAMONLINE_TOWNS_PATH" envDefault:"/api/v1/crm/towns"`
	UmramonlineBranchesPath            string `env:"UMRAMONLINE_BRANCHES_PATH" envDefault:"/api/v1/crm/branches"`
	UmramonlineTaskSMSPath             string `env:"UMRAMONLINE_TASK_SMS_PATH" envDefault:"/api/v1/crm/tasks/sms-created"`
	UmramonlineDashboardVehicleEntryPath string `env:"UMRAMONLINE_DASHBOARD_VEHICLE_ENTRY_PATH" envDefault:"/api/v1/crm/dashboard/vehicle-entry-count"`
	UmramonlineDashboardTotalAmountPath  string `env:"UMRAMONLINE_DASHBOARD_TOTAL_AMOUNT_PATH" envDefault:"/api/v1/crm/dashboard/total-amount"`
	UmramonlineDashboardLoadedCreditPath string `env:"UMRAMONLINE_DASHBOARD_LOADED_CREDIT_PATH" envDefault:"/api/v1/crm/dashboard/loaded-credit"`
	UmramonlineTimeoutSeconds          int    `env:"UMRAMONLINE_TIMEOUT_SECONDS" envDefault:"10"`
	CustomerSyncDailyAt                string `env:"CUSTOMER_SYNC_DAILY_AT" envDefault:"03:00"`
	CustomerSyncCron                   string `env:"CUSTOMER_SYNC_CRON" envDefault:"0 3 * * *"`
	CustomerSyncBatchSize              int    `env:"CUSTOMER_SYNC_BATCH_SIZE" envDefault:"500"`
	CustomerSyncUmramonlineDatabaseDSN string `env:"CUSTOMER_SYNC_UMRAMONLINE_DATABASE_DSN" envDefault:"root:root@tcp(127.0.0.1:33007)/umramdb?charset=utf8mb4&parseTime=True&loc=Local"`
	DatabaseDSN                        string `env:"DATABASE_DSN"`
	SessionTokenSecret                 string `env:"SESSION_TOKEN_SECRET" envDefault:"dev-session-token-secret-change-me"`
	AccessTokenTTLMinutes              int    `env:"ACCESS_TOKEN_TTL_MINUTES" envDefault:"15"`
	RefreshTokenTTLDays                int    `env:"REFRESH_TOKEN_TTL_DAYS" envDefault:"30"`
	AuthCookieSecure                   bool   `env:"AUTH_COOKIE_SECURE" envDefault:"false"`
	AuthCookieSameSite                 string `env:"AUTH_COOKIE_SAME_SITE" envDefault:"Lax"`
	ShutdownTimeoutSeconds             int    `env:"SHUTDOWN_TIMEOUT_SECONDS" envDefault:"10"`
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
