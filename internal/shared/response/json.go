package response

import "github.com/gofiber/fiber/v2"

type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *fiber.Ctx, status int, message string, errors any) error {
	return c.Status(status).JSON(Envelope{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}
