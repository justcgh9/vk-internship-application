package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidInput       = errors.New("invalid input format")
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (*models.User, string, error)
	Login(ctx context.Context, username, password string) (token string, err error)
	VerifyToken(token string) (int64, error)
	GetUser(ctx context.Context, id int64) (*models.User, error)
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
	log := logger.
		FromContext(ctx).
		With("component", "service", "method", "Register")

	if len(username) < 3 || len(password) < 6 || strings.TrimSpace(username) == "" {
		log.Warn("invalid registration input", slog.String("username", username))
		return nil, "", ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", slog.String("err", err.Error()))
		return nil, "", err
	}

	user, err := s.userRepo.CreateUser(ctx, username, string(hash))
	if err != nil {
		log.Error("failed to create user", slog.String("err", err.Error()))
		return nil, "", err
	}

	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		log.Error("failed to generate token", slog.String("err", err.Error()))
		return nil, "", err
	}

	log.Info("user registered successfully", slog.Int64("user_id", user.ID))
	return user, token, nil
}

func (s *service) Login(ctx context.Context, username, password string) (string, error) {
	log := logger.
		FromContext(ctx).
		With("component", "service", "method", "Login")

	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		log.Warn("user not found", slog.String("username", username))
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Warn("invalid password", slog.Int64("user_id", user.ID))
		return "", ErrInvalidCredentials
	}

	token, err := s.tokenManager.GenerateToken(user.ID)
	if err != nil {
		log.Error("failed to generate token", slog.Int64("user_id", user.ID), slog.String("err", err.Error()))
		return "", err
	}

	log.Info("login successful", slog.Int64("user_id", user.ID))
	return token, nil
}

func (s *service) GetUser(ctx context.Context, id int64) (*models.User, error) {
	log := logger.
		FromContext(ctx).
		With("component", "service", "method", "GetUser", "user_id", id)

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		log.Warn("user not found")
		return nil, ErrInvalidCredentials
	}

	log.Debug("user found")
	return user, nil
}

func (s *service) VerifyToken(token string) (int64, error) {
	return s.tokenManager.ParseToken(token)
}
