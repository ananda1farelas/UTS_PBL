package service

import (
	"errors"
	"fmt"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/repository"
)

var (
	ErrDuplicateEnrollment = errors.New("student already enrolled in this course for this academic year")
	ErrQuotaFull           = errors.New("course quota is full")
	ErrSKSExceeded         = errors.New("sks limit exceeded")
	ErrForbiddenEnrollment = errors.New("enrollment does not belong to this student")
)

// SKSExceededError carries the remaining SKS so the handler can include it
// in the error message, per the spec ("sertakan pesan yang menyebut sisa SKS").
type SKSExceededError struct {
	SisaSKS int
}

func (e *SKSExceededError) Error() string {
	return fmt.Sprintf("sks limit exceeded, sisa sks: %d", e.SisaSKS)
}

type EnrollmentService struct {
	enrollmentRepo *repository.EnrollmentRepository
	courseRepo     *repository.CourseRepository
	studentRepo    *repository.StudentRepository
}

func NewEnrollmentService(enrollmentRepo *repository.EnrollmentRepository, courseRepo *repository.CourseRepository, studentRepo *repository.StudentRepository) *EnrollmentService {
	return &EnrollmentService{enrollmentRepo: enrollmentRepo, courseRepo: courseRepo, studentRepo: studentRepo}
}

// Create enrolls a student into a course for a given academic year,
// enforcing every business rule from the spec inside one transaction.
func (s *EnrollmentService) Create(studentID, courseID int, tahunAkademik string) (*model.Enrollment, error) {
	student, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return nil, err
	}

	tx, err := s.enrollmentRepo.BeginTx()
	if err != nil {
		return nil, err
	}

	course, err := s.courseRepo.LockForUpdate(tx, courseID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	duplicate, err := s.enrollmentRepo.ExistsForStudentCourseYear(tx, studentID, courseID, tahunAkademik)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if duplicate {
		tx.Rollback()
		return nil, ErrDuplicateEnrollment
	}

	terisi, err := s.enrollmentRepo.CountForCourse(tx, courseID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if terisi >= course.Kuota {
		tx.Rollback()
		return nil, ErrQuotaFull
	}

	currentSKS, err := s.enrollmentRepo.SumSKSForStudentYear(tx, studentID, tahunAkademik)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	maxSKS := student.MaxSKS()
	if currentSKS+course.SKS > maxSKS {
		tx.Rollback()
		return nil, &SKSExceededError{SisaSKS: maxSKS - currentSKS}
	}

	enrollment, err := s.enrollmentRepo.Create(tx, studentID, courseID, tahunAkademik)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return enrollment, nil
}

// Delete cancels an enrollment, but only if it belongs to studentID.
func (s *EnrollmentService) Delete(enrollmentID, studentID int) error {
	enrollment, err := s.enrollmentRepo.FindByID(enrollmentID)
	if err != nil {
		return err
	}

	if enrollment.StudentID != studentID {
		return ErrForbiddenEnrollment
	}

	return s.enrollmentRepo.Delete(enrollmentID)
}
