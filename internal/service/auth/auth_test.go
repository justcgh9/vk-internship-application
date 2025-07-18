package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/service/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// --- Mocks ---

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CreateUser(ctx context.Context, username, passwordHash string) (*models.User, error) {
	args := m.Called(ctx, username, passwordHash)
	if usr := args.Get(0); usr != nil {
		return usr.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if usr := args.Get(0); usr != nil {
		return usr.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	if usr := args.Get(0); usr != nil {
		return usr.(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// --- Helpers ---

func newTokenManager() *auth.TokenManager {
	return auth.NewTokenManager("secret", time.Hour)
}

// --- Tests: Register ---

func TestRegister_Success(t *testing.T) {
	repo := new(mockUserRepo)
	tm := newTokenManager()
	svc := auth.New(repo, tm)

	username := "alice"
	password := "securepass"

	repo.On("CreateUser", mock.Anything, username, mock.Anything).Return(&models.User{
		ID:       1,
		Username: username,
	}, nil)

	ctx := context.Background()
	usr, token, err := svc.Register(ctx, username, password)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), usr.ID)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestRegister_InvalidInput(t *testing.T) {
	repo := new(mockUserRepo)
	tm := newTokenManager()
	svc := auth.New(repo, tm)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"Empty username", "", "password123"},
		{"Short password", "user", "123"},
		{"Short username", "ab", "validpassword"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := svc.Register(context.Background(), tt.username, tt.password)
			assert.ErrorIs(t, err, auth.ErrInvalidInput)
		})
	}
}

func TestRegister_RepoError(t *testing.T) {
	repo := new(mockUserRepo)
	tm := newTokenManager()
	svc := auth.New(repo, tm)

	repo.On("CreateUser", mock.Anything, "bob", mock.Anything).Return(nil, errors.New("db error"))

	_, _, err := svc.Register(context.Background(), "bob", "password123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

// --- Tests: Login ---

func TestLogin_Success(t *testing.T) {
	repo := new(mockUserRepo)
	tm := newTokenManager()
	svc := auth.New(repo, tm)

	hash, _ := bcrypt.GenerateFromPassword([]byte("mypassword"), bcrypt.DefaultCost)

	repo.On("GetUserByUsername", mock.Anything, "john").Return(&models.User{
		ID:           2,
		Username:     "john",
		PasswordHash: string(hash),
	}, nil)

	token, err := svc.Login(context.Background(), "john", "mypassword")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := new(mockUserRepo)
	tm := newTokenManager()
	svc := auth.New(repo, tm)

	hash, _ := bcrypt.GenerateFromPassword([]byte("rightpass"), bcrypt.DefaultCost)

	repo.On("GetUserByUsername", mock.Anything, "john").Return(&models.User{
		ID:           3,
		Username:     "john",
		PasswordHash: string(hash),
	}, nil)

	_, err := svc.Login(context.Background(), "john", "wrongpass")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	tm := newTokenManager()
	svc := auth.New(repo, tm)

	repo.On("GetUserByUsername", mock.Anything, "ghost").Return(nil, errors.New("not found"))

	_, err := svc.Login(context.Background(), "ghost", "irrelevant")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

// --- Tests: GetUser ---

func TestGetUser_Success(t *testing.T) {
	repo := new(mockUserRepo)
	tm := newTokenManager()
	svc := auth.New(repo, tm)

	repo.On("GetUserByID", mock.Anything, int64(42)).Return(&models.User{
		ID:       42,
		Username: "bob",
	}, nil)

	usr, err := svc.GetUser(context.Background(), 42)
	assert.NoError(t, err)
	assert.Equal(t, "bob", usr.Username)
}

func TestGetUser_Error(t *testing.T) {
	repo := new(mockUserRepo)
	tm := newTokenManager()
	svc := auth.New(repo, tm)

	repo.On("GetUserByID", mock.Anything, int64(99)).Return(nil, errors.New("db error"))

	_, err := svc.GetUser(context.Background(), 99)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

// --- Tests: VerifyToken ---

func TestVerifyToken_Success(t *testing.T) {
	tm := newTokenManager()
	token, err := tm.GenerateToken(123)
	assert.NoError(t, err)

	userID, err := tm.ParseToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), userID)
}

func TestVerifyToken_Invalid(t *testing.T) {
	tm := newTokenManager()

	_, err := tm.ParseToken("invalid.jwt.token")
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}
