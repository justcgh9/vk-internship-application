package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	authHandlers "github.com/justcgh9/vk-internship-application/internal/http/handlers/auth"
	"github.com/justcgh9/vk-internship-application/internal/models"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, username, password string) (*models.User, string, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(*models.User), args.String(1), args.Error(2)
}

func (m *mockAuthService) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func (m *mockAuthService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	panic("not needed")
}

func (m *mockAuthService) VerifyToken(token string) (int64, error) {
	args := m.Called(token)
	return args.Get(0).(int64), args.Error(1)
}

func TestRegister_Success(t *testing.T) {
	authSvc := new(mockAuthService)
	validate := validator.New()
	handler := authHandlers.New(authSvc, validate)

	body := map[string]string{
		"username": "newuser",
		"password": "securepass123",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
	w := httptest.NewRecorder()

	user := &models.User{ID: 1, Username: "newuser"}
	token := "mocked-token"

	authSvc.
		On("Register", mock.Anything, "newuser", "securepass123").
		Return(user, token, nil)

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var out map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)
	require.Equal(t, token, out["token"])
	require.Equal(t, "newuser", out["user"].(map[string]interface{})["username"])
}

func TestRegister_InvalidJSON(t *testing.T) {
	authSvc := new(mockAuthService)
	handler := authHandlers.New(authSvc, validator.New())

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Result().StatusCode)
}

func TestRegister_ValidationError(t *testing.T) {
	authSvc := new(mockAuthService)
	handler := authHandlers.New(authSvc, validator.New())

	body := map[string]string{
		"username": "ab",  // too short
		"password": "123", // too short
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Result().StatusCode)
}

func TestRegister_ServiceError(t *testing.T) {
	authSvc := new(mockAuthService)
	handler := authHandlers.New(authSvc, validator.New())

	body := map[string]string{
		"username": "validuser",
		"password": "validpass123",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
	w := httptest.NewRecorder()

	authSvc.
		On("Register", mock.Anything, "validuser", "validpass123").
		Return(&models.User{}, "", errors.New("registration failed"))

	handler.Register(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestLogin_Success(t *testing.T) {
	authSvc := new(mockAuthService)
	handler := authHandlers.New(authSvc, validator.New())

	body := map[string]string{
		"username": "existinguser",
		"password": "correctpass",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
	w := httptest.NewRecorder()

	authSvc.
		On("Login", mock.Anything, "existinguser", "correctpass").
		Return("jwt-token", nil)

	handler.Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out map[string]string
	err := json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)
	require.Equal(t, "jwt-token", out["token"])
}

func TestLogin_InvalidJSON(t *testing.T) {
	authSvc := new(mockAuthService)
	handler := authHandlers.New(authSvc, validator.New())

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Result().StatusCode)
}

func TestLogin_ValidationError(t *testing.T) {
	authSvc := new(mockAuthService)
	handler := authHandlers.New(authSvc, validator.New())

	body := map[string]string{
		"username": "", // missing
		"password": "",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Result().StatusCode)
}

func TestLogin_Unauthorized(t *testing.T) {
	authSvc := new(mockAuthService)
	handler := authHandlers.New(authSvc, validator.New())

	body := map[string]string{
		"username": "wronguser",
		"password": "wrongpass",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
	w := httptest.NewRecorder()

	authSvc.
		On("Login", mock.Anything, "wronguser", "wrongpass").
		Return("", errors.New("invalid credentials"))

	handler.Login(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}
