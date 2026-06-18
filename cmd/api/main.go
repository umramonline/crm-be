package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	authapp "github.com/umran/new.crm/backend/internal/auth/application"
	authhttp "github.com/umran/new.crm/backend/internal/auth/infrastructure/http"
	"github.com/umran/new.crm/backend/internal/infrastructure/config"
	httpserver "github.com/umran/new.crm/backend/internal/infrastructure/http"
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
		OTPRequestPath:    cfg.UmramonlineOTPRequestPath,
		OTPVerifyPath:     cfg.UmramonlineOTPVerifyPath,
		PasswordLoginPath: cfg.UmramonlinePasswordPath,
		Timeout:           cfg.UmramonlineTimeout(),
	})
	otpRequestService := authapp.NewOTPRequestService(umramonlineClient)
	sessionTokenService := authapp.NewSessionTokenService(cfg.SessionTokenSecret)
	otpHandler := authhttp.NewOTPHandler(otpRequestService, sessionTokenService, authhttp.SessionConfig{
		AccessTTL:      cfg.AccessTokenTTL(),
		RefreshTTL:     cfg.RefreshTokenTTL(),
		CookieSecure:   cfg.AuthCookieSecure,
		CookieSameSite: cfg.AuthCookieSameSite,
	})

	server := httpserver.NewServer(httpserver.Config{
		Addr:                 cfg.Addr(),
		CORSAllowedOrigins:   cfg.CORSAllowedOrigins,
		CORSAllowCredentials: cfg.CORSAllowCredentials,
	}, otpHandler)

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
