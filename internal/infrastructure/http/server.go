package http

import (
	"github.com/gofiber/fiber/v2"

	app "github.com/umran/new.crm/backend/internal/application/greeting"
	"github.com/umran/new.crm/backend/internal/infrastructure/http/handler"
)

type Server struct {
	addr string
	app  *fiber.App
}

func NewServer(addr string) *Server {
	greetingService := app.NewService()
	helloHandler := handler.NewHelloHandler(greetingService)

	fiberApp := fiber.New()
	fiberApp.Get("/", helloHandler.Handle)

	return &Server{
		addr: addr,
		app:  fiberApp,
	}
}

func (s *Server) Run() error {
	return s.app.Listen(s.addr)
}
