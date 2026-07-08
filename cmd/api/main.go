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
	customersync "github.com/umran/new.crm/backend/internal/customer/sync"
	followupapp "github.com/umran/new.crm/backend/internal/followup/application"
	followuphttp "github.com/umran/new.crm/backend/internal/followup/infrastructure/http"
	followuppersistence "github.com/umran/new.crm/backend/internal/followup/infrastructure/persistence"
	followupstorage "github.com/umran/new.crm/backend/internal/followup/infrastructure/storage"
	"github.com/umran/new.crm/backend/internal/infrastructure/config"
	httpserver "github.com/umran/new.crm/backend/internal/infrastructure/http"
	dbpersistence "github.com/umran/new.crm/backend/internal/infrastructure/persistence"
	taskapp "github.com/umran/new.crm/backend/internal/task/application"
	taskhttp "github.com/umran/new.crm/backend/internal/task/infrastructure/http"
	taskpersistence "github.com/umran/new.crm/backend/internal/task/infrastructure/persistence"
	taskumramonline "github.com/umran/new.crm/backend/internal/task/infrastructure/umramonline"
	"github.com/umran/new.crm/backend/internal/umramonline"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	umramonlineClient := umramonline.NewClient(umramonline.Config{
		BaseURL:                 cfg.UmramonlineBaseURL,
		APIKey:                  cfg.UmramonlineAPIKey,
		APIToken:                cfg.UmramonlineAPIToken,
		OTPRequestPath:          cfg.UmramonlineOTPRequestPath,
		OTPVerifyPath:           cfg.UmramonlineOTPVerifyPath,
		PasswordLoginPath:       cfg.UmramonlinePasswordPath,
		UserRolesPath:           cfg.UmramonlineUserRolesPath,
		CustomersPath:           cfg.UmramonlineCustomersPath,
		CustomerSearchPath:      cfg.UmramonlineCustomerSearchPath,
		CustomerPhoneExistsPath: cfg.UmramonlineCustomerPhoneExistsPath,
		ZonesPath:               cfg.UmramonlineZonesPath,
		CitiesPath:              cfg.UmramonlineCitiesPath,
		TownsPath:               cfg.UmramonlineTownsPath,
		BranchesPath:            cfg.UmramonlineBranchesPath,
		TaskSMSPath:             cfg.UmramonlineTaskSMSPath,
		Timeout:                 cfg.UmramonlineTimeout(),
	})
	otpRequestService := authapp.NewOTPRequestService(umramonlineClient)
	sessionTokenService := authapp.NewSessionTokenService(cfg.SessionTokenSecret)
	authorizationProvider := authzumramonline.NewProvider(umramonlineClient)
	var permissionRepository authzapp.PermissionRepository
	var moduleRepository authzapp.ModuleRepository
	var customerRepository customerapp.CustomerRepository
	var db *gorm.DB

	db, err = dbpersistence.OpenMySQL(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	if err := authzpersistence.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	if err := customerpersistence.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	if err := taskpersistence.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	if err := followuppersistence.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	if err := authzpersistence.SeedAuthorization(db); err != nil {
		log.Fatal(err)
	}

	if err := authzpersistence.SeedCustomers(db); err != nil {
		log.Fatal(err)
	}

	if err := authzpersistence.SeedTasks(db); err != nil {
		log.Fatal(err)
	}

	authorizationRepository := authzpersistence.NewRepository(db)
	permissionRepository = authorizationRepository
	moduleRepository = authorizationRepository
	customerRepository = customerpersistence.NewRepository(db)

	authorizationService := authzapp.NewService(authorizationProvider, permissionRepository, moduleRepository)
	otpHandler := authhttp.NewOTPHandler(otpRequestService, sessionTokenService, authhttp.SessionConfig{
		AccessTTL:      cfg.AccessTokenTTL(),
		RefreshTTL:     cfg.RefreshTokenTTL(),
		CookieSecure:   cfg.AuthCookieSecure,
		CookieSameSite: cfg.AuthCookieSameSite,
	})
	otpHandler.SetAuthorizationService(authzhttp.NewSessionAdapter(authorizationService))
	authorizationHandler := authzhttp.NewHandler(authorizationService)
	customerService := customerapp.NewService(customerumramonline.NewProvider(umramonlineClient), customerRepository)
	customerHandler := customerhttp.NewHandler(customerService)
	taskRepository := taskpersistence.NewRepository(db)
	taskProvider := taskumramonline.NewProvider(umramonlineClient)
	taskService := taskapp.NewService(taskProvider, taskRepository)
	taskHandler := taskhttp.NewHandler(taskService, taskProvider)
	followUpRepository := followuppersistence.NewRepository(db)
	followUpStorage := followupstorage.NewLocalImageStorage("storage/follow-ups", "/storage/follow-ups")
	followUpService := followupapp.NewService(followUpRepository, followUpStorage)
	followUpHandler := followuphttp.NewHandler(followUpService)
	authRequired := authzhttp.RequirePermission(authorizationService, sessionTokenService, authzhttp.AuthMiddlewareConfig{})

	server := httpserver.NewServer(httpserver.Config{
		Addr:                 cfg.Addr(),
		CORSAllowedOrigins:   cfg.CORSAllowedOrigins,
		CORSAllowCredentials: cfg.CORSAllowCredentials,
	}, otpHandler, authorizationHandler, customerHandler, taskHandler, followUpHandler, authRequired)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	umramonlineDB, err := dbpersistence.OpenMySQL(cfg.CustomerSyncUmramonlineDatabaseDSN)
	if err != nil {
		log.Fatalf("open umramonline database for customer sync: %v", err)
	}

	customersync.StartDailyScheduler(ctx, customersync.SchedulerConfig{
		SourceDB:  umramonlineDB,
		TargetDB:  db,
		BatchSize: cfg.CustomerSyncBatchSize,
		DailyAt:   cfg.CustomerSyncDailyAt,
		CronExpr:  cfg.CustomerSyncCron,
		Logger:    log.Default(),
	})
	log.Printf("customer sync scheduler enabled cron=%q daily_at_fallback=%s", cfg.CustomerSyncCron, cfg.CustomerSyncDailyAt)

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
