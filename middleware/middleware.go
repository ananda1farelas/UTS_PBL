package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ananda1farelas/latihan-fiber/UTS/config"
	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
)

const (
	LocalUserID = "user_id"
	LocalEmail  = "email"
	LocalRole   = "role"
)

func Auth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return helper.Error(c, fiber.StatusUnauthorized, "Token tidak ditemukan")
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims, err := helper.ParseToken(cfg.JWTSecret, tokenString)
		if err != nil {
			return helper.Error(c, fiber.StatusUnauthorized, "Token tidak valid atau kedaluwarsa")
		}

		c.Locals(LocalUserID, claims.UserID)
		c.Locals(LocalEmail, claims.Email)
		c.Locals(LocalRole, claims.Role)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals(LocalRole).(string)

		for _, allowed := range roles {
			if role == allowed {
				return c.Next()
			}
		}

		return helper.Error(c, fiber.StatusForbidden, "Anda tidak memiliki akses ke resource ini")
	}
}

type loginAttempt struct {
	count     int
	windowEnd time.Time
}

type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
	max      int
	window   time.Duration
}

func NewLoginRateLimiter(max int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		attempts: make(map[string]*loginAttempt),
		max:      max,
		window:   window,
	}
}

func (l *LoginRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, exists := l.attempts[key]
	now := time.Now()

	if !exists || now.After(a.windowEnd) {
		l.attempts[key] = &loginAttempt{count: 0, windowEnd: now.Add(l.window)}
		return true
	}

	return a.count < l.max
}

func (l *LoginRateLimiter) RegisterFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, exists := l.attempts[key]
	now := time.Now()

	if !exists || now.After(a.windowEnd) {
		l.attempts[key] = &loginAttempt{count: 1, windowEnd: now.Add(l.window)}
		return
	}

	a.count++
}

func (l *LoginRateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.IP()
		if !l.Allow(key) {
			return helper.Error(c, fiber.StatusTooManyRequests, "Terlalu banyak percobaan login, coba lagi nanti")
		}
		return c.Next()
	}
}
