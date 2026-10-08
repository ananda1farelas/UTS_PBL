package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/repository"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/service"
	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
)

type CourseHandler struct {
	courseService *service.CourseService
}

func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

// List handles GET /api/v1/courses (all authenticated roles).
func (h *CourseHandler) List(c *fiber.Ctx) error {
	filter := repository.CourseFilter{
		Semester:  c.Query("semester"),
		Search:    c.Query("search"),
		Available: c.Query("available") == "true",
	}

	courses, err := h.courseService.List(filter)
	if err != nil {
		return helper.Error(c, fiber.StatusInternalServerError, "Gagal mengambil data mata kuliah")
	}

	if courses == nil {
		courses = []model.Course{}
	}

	return helper.Success(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", courses)
}
