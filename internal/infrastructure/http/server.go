package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	app "github.com/umran/new.crm/backend/internal/application/greeting"
	authhttp "github.com/umran/new.crm/backend/internal/auth/infrastructure/http"
	authzhttp "github.com/umran/new.crm/backend/internal/authorization/infrastructure/http"
	customerhttp "github.com/umran/new.crm/backend/internal/customer/infrastructure/http"
	followuphttp "github.com/umran/new.crm/backend/internal/followup/infrastructure/http"
	dashboardhttp "github.com/umran/new.crm/backend/internal/dashboard/infrastructure/http"
	"github.com/umran/new.crm/backend/internal/infrastructure/http/handler"
	iettshttp "github.com/umran/new.crm/backend/internal/ietts/infrastructure/http"
	taskhttp "github.com/umran/new.crm/backend/internal/task/infrastructure/http"
)

type Server struct {
	addr string
	app  *fiber.App
}

type Config struct {
	Addr                 string
	CORSAllowedOrigins   string
	CORSAllowCredentials bool
}

func NewServer(config Config,
	otpHandler *authhttp.OTPHandler,
	authorizationHandler *authzhttp.Handler,
	customerHandler *customerhttp.Handler,
	taskHandler *taskhttp.Handler,
	followUpHandler *followuphttp.Handler,
	iettsHandler *iettshttp.Handler,
	dashboardHandler *dashboardhttp.Handler,
	authRequired fiber.Handler) *Server {
	greetingService := app.NewService()
	helloHandler := handler.NewHelloHandler(greetingService)

	fiberApp := fiber.New()
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     config.CORSAllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowCredentials: config.CORSAllowCredentials,
	}))

	fiberApp.Get("/", helloHandler.Handle)
	fiberApp.Static("/storage/follow-ups", "storage/follow-ups")

	apiV1 := fiberApp.Group("/api/v1")
	apiV1.Post("/auth/otp/request", otpHandler.RequestOTP)
	apiV1.Post("/auth/otp/verify", otpHandler.VerifyOTP)
	apiV1.Post("/auth/password/login", otpHandler.LoginWithPassword)
	apiV1.Post("/auth/refresh", otpHandler.RefreshSession)
	apiV1.Post("/auth/logout", otpHandler.Logout)
	apiV1.Get("/auth/session", otpHandler.Session)
	authorizationHandler.RegisterRoutes(apiV1, authRequired)
	customerHandler.RegisterRoutes(apiV1, authRequired)
	taskHandler.RegisterRoutes(apiV1, authRequired)
	followUpHandler.RegisterRoutes(apiV1, authRequired)
	iettsHandler.RegisterRoutes(apiV1, authRequired)
	dashboardHandler.RegisterRoutes(apiV1, authRequired)

	return &Server{
		addr: config.Addr,
		app:  fiberApp,
	}
}

func (s *Server) Run() error {
	return s.app.Listen(s.addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
