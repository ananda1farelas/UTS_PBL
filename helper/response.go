package helper

import "github.com/gofiber/fiber/v2"

// Meta carries pagination info for list endpoints.
type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

// Success writes the uniform success envelope used across every endpoint.
func Success(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// SuccessWithMeta is Success plus a pagination meta object, used by list
// endpoints such as GET /students.
func SuccessWithMeta(c *fiber.Ctx, status int, message string, data interface{}, meta Meta) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

// Error writes the uniform error envelope (no "errors" field).
func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// ValidationError writes a 422 response with a per-field errors map, matching
// the format mandated by the spec.
func ValidationError(c *fiber.Ctx, errors map[string][]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"success": false,
		"message": "Validasi gagal",
		"errors":  errors,
	})
}
