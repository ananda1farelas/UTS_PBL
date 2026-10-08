package service

import (
	"errors"
	"time"

	"github.com/ananda1farelas/latihan-fiber/UTS/app/model"
	"github.com/ananda1farelas/latihan-fiber/UTS/app/repository"
	"github.com/ananda1farelas/latihan-fiber/UTS/config"
	"github.com/ananda1farelas/latihan-fiber/UTS/helper"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	cfg         *config.Config
	userRepo    *repository.UserRepository
	studentRepo *repository.StudentRepository
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository, studentRepo *repository.StudentRepository) *AuthService {
	return &AuthService{cfg: cfg, userRepo: userRepo, studentRepo: studentRepo}
}

// LoginResult is what AuthService.Login returns on success.
type LoginResult struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64 // seconds
	User        *model.User
}

// Login verifies credentials and issues a JWT. It returns
// ErrInvalidCredentials for both "user not found" and "wrong password" so
// callers never leak which one failed.
func (s *AuthService) Login(email, password string) (*LoginResult, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !helper.CheckPassword(user.Password, password) {
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := helper.GenerateToken(s.cfg.JWTSecret, s.cfg.JWTExpiresIn, user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
		User:        user,
	}, nil
}

// Me returns the logged-in user, plus their student profile when they hold
// the mahasiswa role.
func (s *AuthService) Me(userID int) (*model.User, *model.Student, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, err
	}

	if user.Role != model.RoleMahasiswa {
		return user, nil, nil
	}

	student, err := s.studentRepo.FindByUserID(userID)
	if err != nil {
		return nil, nil, err
	}

	return user, student, nil
}
