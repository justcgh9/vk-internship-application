package auth

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidInput       = errors.New("invalid input format")
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (*models.User, string, error)
	Login(ctx context.Context, username, password string) (token string, err error)
	VerifyToken(token string) (int64, error)
}

type service struct {
	userRepo     storage.UserRepository
	tokenManager *TokenManager
}

func New(userRepo storage.UserRepository, tm *TokenManager) AuthService {
	return &service{
		userRepo:     userRepo,
		tokenManager: tm,
	}
}

func (s *service) Register(ctx context.Context, username, password string) (*models.User, string, error) {
	if len(username) < 3 || len(password) < 6 || strings.TrimSpace(username) == "" {
		return nil, "", ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	usr, err := s.userRepo.CreateUser(ctx, username, string(hash))
	if err != nil {
		return nil, "", err
	}

	token, err := s.tokenManager.GenerateToken(usr.ID)
	if err != nil {
		return nil, "", err
	}

	return usr, token, err
}

func (s *service) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.tokenManager.GenerateToken(user.ID)
	return token, err
}

func (s *service) VerifyToken(token string) (int64, error) {
	return s.tokenManager.ParseToken(token)
}
