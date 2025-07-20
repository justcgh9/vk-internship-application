package listings_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/justcgh9/vk-internship-application/internal/http/handlers/listings"
	"github.com/justcgh9/vk-internship-application/internal/http/middleware"
	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockAuthService) Register(ctx context.Context, username, password string) (*models.User, string, error) {
	panic("not used in this test")
}
func (m *mockAuthService) Login(ctx context.Context, username, password string) (string, error) {
	panic("not used in this test")
}
func (m *mockAuthService) VerifyToken(token string) (int64, error) {
	panic("not used in this test")
}

type mockListingService struct {
	mock.Mock
}

func (m *mockListingService) Create(ctx context.Context, l *models.Listing) (*models.Listing, error) {
	args := m.Called(ctx, l)
	return args.Get(0).(*models.Listing), args.Error(1)
}

func (m *mockListingService) List(ctx context.Context, f storage.ListFilter) ([]*models.ListingWithAuthor, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]*models.ListingWithAuthor), args.Error(1)
}

func TestCreateListing(t *testing.T) {
	validate := validator.New()

	authSvc := new(mockAuthService)
	listingSvc := new(mockListingService)

	h := listings.New(authSvc, listingSvc, validate)

	userID := int64(123)
	input := listings.CreateListingRequest{
		Title:       "Test Listing",
		Description: "A valid description for the listing",
		ImageURL:    "https://i.pinimg.com/474x/bd/a8/0e/bda80e9324bd6d5c83b84b6eac5a1e5d.jpg",
		Price:       123.45,
	}
	body, _ := json.Marshal(input)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserID(context.Background(), userID))

	w := httptest.NewRecorder()

	createdListing := &models.Listing{
		ID:          1,
		Title:       input.Title,
		Description: input.Description,
		ImageURL:    input.ImageURL,
		Price:       input.Price,
		UserID:      userID,
		CreatedAt:   time.Now(),
	}

	listingSvc.
		On("Create", mock.Anything, mock.MatchedBy(func(l *models.Listing) bool {
			return l.Title == input.Title && l.UserID == userID
		})).
		Return(createdListing, nil)

	authSvc.
		On("GetUser", mock.Anything, userID).
		Return(&models.User{ID: userID, Username: "tester"}, nil)

	h.CreateListing(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var out models.ListingWithAuthor
	err := json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)

	require.Equal(t, createdListing.ID, out.ID)
	require.Equal(t, "tester", out.AuthorLogin)
	require.True(t, out.IsOwned)
}
