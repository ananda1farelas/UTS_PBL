package helper

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// GlobalErrorHandler is registered as Fiber's app-wide ErrorHandler. It keeps
// every unexpected error (including panics recovered by the recover
// middleware) inside the same JSON envelope as the rest of the API, and
// never leaks a stack trace, per the spec's 500 requirement.
func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "Terjadi kesalahan pada server"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
		message = fiberErr.Message
	}

	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// NotFoundHandler responds to any route that doesn't match a registered
// endpoint.
func NotFoundHandler(c *fiber.Ctx) error {
	return Error(c, fiber.StatusNotFound, "Endpoint tidak ditemukan")
}
