package route

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/handler"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/repository"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/service"
	"github.com/ananda1farelas/latihan-fiber/UTS/config"
	"github.com/ananda1farelas/latihan-fiber/UTS/middleware"
)

// Setup wires repositories -> services -> handlers and registers every
// endpoint from the spec under /api/v1.
func Setup(app *fiber.App, db *sql.DB, cfg *config.Config) {
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	courseRepo := repository.NewCourseRepository(db)
	enrollmentRepo := repository.NewEnrollmentRepository(db)

	authService := service.NewAuthService(cfg, userRepo, studentRepo)
	studentService := service.NewStudentService(studentRepo, userRepo)
	courseService := service.NewCourseService(courseRepo)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, courseRepo, studentRepo)

	loginLimiter := middleware.NewLoginRateLimiter(cfg.LoginRateLimitMax, cfg.LoginRateLimitWindow)

	authHandler := handler.NewAuthHandler(authService, loginLimiter)
	studentHandler := handler.NewStudentHandler(studentService)
	courseHandler := handler.NewCourseHandler(courseService)
	enrollmentHandler := handler.NewEnrollmentHandler(enrollmentService, studentRepo)

	auth := middleware.Auth(cfg)
	adminOnly := middleware.RequireRole(model.RoleAdmin)
	mahasiswaOnly := middleware.RequireRole(model.RoleMahasiswa)

	v1 := app.Group("/api/v1")

	// 1. POST /api/v1/auth/login — public
	v1.Post("/auth/login", loginLimiter.Middleware(), authHandler.Login)

	// 2. GET /api/v1/auth/me — any authenticated role
	v1.Get("/auth/me", auth, authHandler.Me)

	// 3-7. /api/v1/students
	v1.Get("/students", auth, adminOnly, studentHandler.List)
	v1.Post("/students", auth, adminOnly, studentHandler.Create)
	v1.Get("/students/:id", auth, studentHandler.Detail) // admin or own data, checked in handler
	v1.Put("/students/:id", auth, adminOnly, studentHandler.Update)
	v1.Delete("/students/:id", auth, adminOnly, studentHandler.Delete)

	// 8. GET /api/v1/courses — any authenticated role
	v1.Get("/courses", auth, courseHandler.List)

	// 9-10. /api/v1/enrollments — mahasiswa only
	v1.Post("/enrollments", auth, mahasiswaOnly, enrollmentHandler.Create)
	v1.Delete("/enrollments/:id", auth, mahasiswaOnly, enrollmentHandler.Delete)
}
