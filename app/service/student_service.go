package service

import (
	"errors"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/repository"
	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
)

var (
	ErrNimTaken   = errors.New("nim already taken")
	ErrEmailTaken = errors.New("email already taken")
)

type StudentService struct {
	studentRepo *repository.StudentRepository
	userRepo    *repository.UserRepository
}

func NewStudentService(studentRepo *repository.StudentRepository, userRepo *repository.UserRepository) *StudentService {
	return &StudentService{studentRepo: studentRepo, userRepo: userRepo}
}

func (s *StudentService) List(filter repository.StudentFilter, p helper.PaginationParams) ([]model.Student, int, error) {
	return s.studentRepo.List(filter, p)
}

// Detail returns a student plus their enrolled courses, total SKS, and SKS cap.
func (s *StudentService) Detail(id int) (*model.StudentDetail, error) {
	student, err := s.studentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	courses, err := s.studentRepo.EnrolledCourses(id)
	if err != nil {
		return nil, err
	}
	if courses == nil {
		courses = []model.Course{}
	}

	totalSKS, err := s.studentRepo.TotalSKS(id)
	if err != nil {
		return nil, err
	}

	return &model.StudentDetail{
		Student:  *student,
		Courses:  courses,
		TotalSKS: totalSKS,
		BatasSKS: student.MaxSKS(),
	}, nil
}

type CreateStudentInput struct {
	NIM         string
	Nama        string
	Email       string
	Prodi       string
	Angkatan    int
	IPKTerakhir float64
}

func (s *StudentService) Create(in CreateStudentInput) (*model.Student, error) {
	nimExists, err := s.studentRepo.NimExists(in.NIM)
	if err != nil {
		return nil, err
	}
	if nimExists {
		return nil, ErrNimTaken
	}

	emailExists, err := s.studentRepo.EmailExists(in.Email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, ErrEmailTaken
	}

	// Initial password is the NIM itself, hashed, per spec.
	hashed, err := helper.HashPassword(in.NIM)
	if err != nil {
		return nil, err
	}

	return s.studentRepo.CreateWithUser(s.userRepo, in.Email, hashed, in.NIM, in.Nama, in.Prodi, in.Angkatan, in.IPKTerakhir)
}

type UpdateStudentInput struct {
	Nama        string
	Prodi       string
	Angkatan    int
	IPKTerakhir float64
}

func (s *StudentService) Update(id int, in UpdateStudentInput) (*model.Student, error) {
	return s.studentRepo.Update(id, in.Nama, in.Prodi, in.Angkatan, in.IPKTerakhir)
}

func (s *StudentService) Delete(id int) error {
	return s.studentRepo.SoftDelete(id)
}
