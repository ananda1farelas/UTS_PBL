package service

import (
	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/repository"
)

type CourseService struct {
	courseRepo *repository.CourseRepository
}

func NewCourseService(courseRepo *repository.CourseRepository) *CourseService {
	return &CourseService{courseRepo: courseRepo}
}

func (s *CourseService) List(filter repository.CourseFilter) ([]model.Course, error) {
	return s.courseRepo.List(filter)
}
