package handler

import (
	"errors"
	"regexp"

	"github.com/gofiber/fiber/v2"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/service"
	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
	"github.com/ananda1farelas/latihan-fiber/UTS/middleware"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type AuthHandler struct {
	authService *service.AuthService
	rateLimiter *middleware.LoginRateLimiter
}

func NewAuthHandler(authService *service.AuthService, rateLimiter *middleware.LoginRateLimiter) *AuthHandler {
	return &AuthHandler{authService: authService, rateLimiter: rateLimiter}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Body request tidak valid")
	}

	errs := map[string][]string{}
	if req.Email == "" {
		errs["email"] = append(errs["email"], "Email wajib diisi")
	} else if !emailRegex.MatchString(req.Email) {
		errs["email"] = append(errs["email"], "Format email tidak valid")
	}
	if req.Password == "" {
		errs["password"] = append(errs["password"], "Password wajib diisi")
	} else if len(req.Password) < 8 {
		errs["password"] = append(errs["password"], "Password minimal 8 karakter")
	}
	if len(errs) > 0 {
		return helper.ValidationError(c, errs)
	}

	result, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.rateLimiter.RegisterFailure(c.IP())
			return helper.Error(c, fiber.StatusUnauthorized, "Email atau password salah")
		}
		return helper.Error(c, fiber.StatusInternalServerError, "Terjadi kesalahan pada server")
	}

	return helper.Success(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"access_token": result.AccessToken,
		"token_type":   result.TokenType,
		"expires_in":   result.ExpiresIn,
		"user": fiber.Map{
			"id":    result.User.ID,
			"email": result.User.Email,
			"role":  result.User.Role,
		},
	})
}

// Me handles GET /api/v1/auth/me.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals(middleware.LocalUserID).(int)

	user, student, err := h.authService.Me(userID)
	if err != nil {
		return helper.Error(c, fiber.StatusUnauthorized, "Token tidak valid atau kedaluwarsa")
	}

	data := fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}

	if student != nil {
		data["students"] = fiber.Map{
			"nim":      student.NIM,
			"nama":     student.Nama,
			"prodi":    student.Prodi,
			"angkatan": student.Angkatan,
		}
	}

	return helper.Success(c, fiber.StatusOK, "Profil berhasil diambil", data)
}
