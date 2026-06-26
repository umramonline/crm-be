package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	authapp "github.com/umran/new.crm/backend/internal/auth/application"
	authhttp "github.com/umran/new.crm/backend/internal/auth/infrastructure/http"
	authzapp "github.com/umran/new.crm/backend/internal/authorization/application"
	authzhttp "github.com/umran/new.crm/backend/internal/authorization/infrastructure/http"
	authzpersistence "github.com/umran/new.crm/backend/internal/authorization/infrastructure/persistence"
	authzumramonline "github.com/umran/new.crm/backend/internal/authorization/infrastructure/umramonline"
	customerapp "github.com/umran/new.crm/backend/internal/customer/application"
	customerhttp "github.com/umran/new.crm/backend/internal/customer/infrastructure/http"
	customerpersistence "github.com/umran/new.crm/backend/internal/customer/infrastructure/persistence"
	customerumramonline "github.com/umran/new.crm/backend/internal/customer/infrastructure/umramonline"
	"github.com/umran/new.crm/backend/internal/infrastructure/config"
	httpserver "github.com/umran/new.crm/backend/internal/infrastructure/http"
	dbpersistence "github.com/umran/new.crm/backend/internal/infrastructure/persistence"
	"github.com/umran/new.crm/backend/internal/umramonline"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	umramonlineClient := umramonline.NewClient(umramonline.Config{
		BaseURL:           cfg.UmramonlineBaseURL,
		APIKey:            cfg.UmramonlineAPIKey,
		APIToken:          cfg.UmramonlineAPIToken,
		OTPRequestPath:    cfg.UmramonlineOTPRequestPath,
		OTPVerifyPath:     cfg.UmramonlineOTPVerifyPath,
		PasswordLoginPath: cfg.UmramonlinePasswordPath,
		UserRolesPath:     cfg.UmramonlineUserRolesPath,
		CustomersPath:     cfg.UmramonlineCustomersPath,
		ZonesPath:         cfg.UmramonlineZonesPath,
		Timeout:           cfg.UmramonlineTimeout(),
	})
	otpRequestService := authapp.NewOTPRequestService(umramonlineClient)
	sessionTokenService := authapp.NewSessionTokenService(cfg.SessionTokenSecret)
	authorizationProvider := authzumramonline.NewProvider(umramonlineClient)
	var permissionRepository authzapp.PermissionRepository
	var moduleRepository authzapp.ModuleRepository
	if cfg.DatabaseDSN != "" {
		db, err := dbpersistence.OpenMySQL(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}

		if err := authzpersistence.AutoMigrate(db); err != nil {
			log.Fatal(err)
		}

		if err := customerpersistence.AutoMigrate(db); err != nil {
			log.Fatal(err)
		}

		if err := authzpersistence.SeedAuthorization(db); err != nil {
			log.Fatal(err)
		}

		if err := authzpersistence.SeedCustomers(db); err != nil {
			log.Fatal(err)
		}

		authorizationRepository := authzpersistence.NewRepository(db)
		permissionRepository = authorizationRepository
		moduleRepository = authorizationRepository
	} else {
		log.Println("DATABASE_DSN is empty; authorization persistence is disabled")
	}

	authorizationService := authzapp.NewService(authorizationProvider, permissionRepository, moduleRepository)
	otpHandler := authhttp.NewOTPHandler(otpRequestService, sessionTokenService, authhttp.SessionConfig{
		AccessTTL:      cfg.AccessTokenTTL(),
		RefreshTTL:     cfg.RefreshTokenTTL(),
		CookieSecure:   cfg.AuthCookieSecure,
		CookieSameSite: cfg.AuthCookieSameSite,
	})
	otpHandler.SetAuthorizationService(authzhttp.NewSessionAdapter(authorizationService))
	authorizationHandler := authzhttp.NewHandler(authorizationService)
	customerService := customerapp.NewService(customerumramonline.NewProvider(umramonlineClient))
	customerHandler := customerhttp.NewHandler(customerService)
	authRequired := authzhttp.RequirePermission(authorizationService, sessionTokenService, authzhttp.AuthMiddlewareConfig{})

	server := httpserver.NewServer(httpserver.Config{
		Addr:                 cfg.Addr(),
		CORSAllowedOrigins:   cfg.CORSAllowedOrigins,
		CORSAllowCredentials: cfg.CORSAllowCredentials,
	}, otpHandler, authorizationHandler, customerHandler, authRequired)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Run()
	}()

	log.Printf("server listening on %s", cfg.Addr())

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	case <-ctx.Done():
		log.Println("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout())
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}

		log.Println("server stopped gracefully")
	}
}
