package handler

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/repository"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/service"
	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
	"github.com/ananda1farelas/latihan-fiber/UTS/middleware"
)

// tahunAkademikRegex matches the "2026/2027-Ganjil" / "2026/2027-Genap" format.
var tahunAkademikRegex = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)

type EnrollmentHandler struct {
	enrollmentService *service.EnrollmentService
	studentRepo       *repository.StudentRepository
}

func NewEnrollmentHandler(enrollmentService *service.EnrollmentService, studentRepo *repository.StudentRepository) *EnrollmentHandler {
	return &EnrollmentHandler{enrollmentService: enrollmentService, studentRepo: studentRepo}
}

type createEnrollmentRequest struct {
	CourseID      int    `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

// currentStudentID resolves the logged-in mahasiswa's student.id from their
// user_id, since the JWT only carries the latter.
func (h *EnrollmentHandler) currentStudentID(c *fiber.Ctx) (int, error) {
	userID, _ := c.Locals(middleware.LocalUserID).(int)
	student, err := h.studentRepo.FindByUserID(userID)
	if err != nil {
		return 0, err
	}
	return student.ID, nil
}

// Create handles POST /api/v1/enrollments (mahasiswa only).
func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	var req createEnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Error(c, fiber.StatusBadRequest, "Body request tidak valid")
	}

	errs := map[string][]string{}
	if req.CourseID == 0 {
		errs["course_id"] = append(errs["course_id"], "course_id wajib diisi")
	}
	if req.TahunAkademik == "" || !tahunAkademikRegex.MatchString(req.TahunAkademik) {
		errs["tahun_akademik"] = append(errs["tahun_akademik"], "tahun_akademik wajib diisi dengan format 2026/2027-Ganjil")
	}
	if len(errs) > 0 {
		return helper.ValidationError(c, errs)
	}

	studentID, err := h.currentStudentID(c)
	if err != nil {
		return helper.Error(c, fiber.StatusForbidden, "Akun ini tidak memiliki profil mahasiswa")
	}

	enrollment, err := h.enrollmentService.Create(studentID, req.CourseID, req.TahunAkademik)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return helper.ValidationError(c, map[string][]string{"course_id": {"Mata kuliah tidak ditemukan"}})
		case errors.Is(err, service.ErrDuplicateEnrollment):
			return helper.Error(c, fiber.StatusConflict, "Mata kuliah ini sudah pernah diambil pada tahun akademik yang sama")
		case errors.Is(err, service.ErrQuotaFull):
			return helper.ValidationError(c, map[string][]string{"course_id": {"Kuota mata kuliah sudah penuh"}})
		default:
			var sksErr *service.SKSExceededError
			if errors.As(err, &sksErr) {
				return helper.ValidationError(c, map[string][]string{
					"course_id": {"Total SKS melebihi batas, sisa SKS Anda: " + strconv.Itoa(sksErr.SisaSKS)},
				})
			}
			return helper.Error(c, fiber.StatusInternalServerError, "Gagal mengambil mata kuliah")
		}
	}

	return helper.Success(c, fiber.StatusCreated, "Mata kuliah berhasil ditambahkan ke KRS", enrollment)
}

// Delete handles DELETE /api/v1/enrollments/{id} (mahasiswa, own enrollment only).
func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return helper.Error(c, fiber.StatusNotFound, "Enrollment tidak ditemukan")
	}

	studentID, err := h.currentStudentID(c)
	if err != nil {
		return helper.Error(c, fiber.StatusForbidden, "Akun ini tidak memiliki profil mahasiswa")
	}

	if err := h.enrollmentService.Delete(id, studentID); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return helper.Error(c, fiber.StatusNotFound, "Enrollment tidak ditemukan")
		case errors.Is(err, service.ErrForbiddenEnrollment):
			return helper.Error(c, fiber.StatusForbidden, "Enrollment ini bukan milik Anda")
		default:
			return helper.Error(c, fiber.StatusInternalServerError, "Gagal membatalkan mata kuliah")
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}
