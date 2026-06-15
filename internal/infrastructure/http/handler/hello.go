package handler

import (
	"github.com/gofiber/fiber/v2"

	app "github.com/umran/new.crm/backend/internal/application/greeting"
)

type HelloHandler struct {
	service *app.Service
}

func NewHelloHandler(service *app.Service) *HelloHandler {
	return &HelloHandler{service: service}
}

func (h *HelloHandler) Handle(c *fiber.Ctx) error {
	greeting := h.service.GetHello()
	return c.Type("text/plain; charset=utf-8").SendString(greeting.Message)
}
