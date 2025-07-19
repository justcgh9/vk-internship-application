package listings_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestListListings_Basic(t *testing.T) {
	validate := validator.New()
	authSvc := new(mockAuthService)
	listingSvc := new(mockListingService)

	h := listings.New(authSvc, listingSvc, validate)

	expected := []*models.ListingWithAuthor{
		{
			ID:          1,
			Title:       "Sample",
			Description: "Something nice",
			ImageURL:    "https://example.com/image.jpg",
			Price:       10.0,
			AuthorLogin: "tester",
			IsOwned:     false,
			CreatedAt:   time.Now(),
		},
	}

	listingSvc.
		On("List", mock.Anything, mock.MatchedBy(func(f storage.ListFilter) bool {
			return f.Limit == 10 && f.Offset == 0 && f.ViewerID == nil
		})).
		Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	h.ListListings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listingsOut []*models.ListingWithAuthor
	err := json.NewDecoder(resp.Body).Decode(&listingsOut)
	require.NoError(t, err)
	require.Len(t, listingsOut, 1)
	require.Equal(t, expected[0].Title, listingsOut[0].Title)
}

func TestListListings_WithViewer(t *testing.T) {
	validate := validator.New()
	authSvc := new(mockAuthService)
	listingSvc := new(mockListingService)

	h := listings.New(authSvc, listingSvc, validate)

	viewerID := int64(42)
	expected := []*models.ListingWithAuthor{
		{
			ID:          2,
			Title:       "Owned listing",
			Description: "Owned by user",
			ImageURL:    "https://example.com/own.jpg",
			Price:       99.9,
			IsOwned:     true,
			CreatedAt:   time.Now(),
		},
	}

	listingSvc.
		On("List", mock.Anything, mock.MatchedBy(func(f storage.ListFilter) bool {
			return f.ViewerID != nil && *f.ViewerID == viewerID
		})).
		Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(middleware.WithUserID(context.Background(), viewerID))
	w := httptest.NewRecorder()

	h.ListListings(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out []*models.ListingWithAuthor
	err := json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.True(t, out[0].IsOwned)
}

func TestListListings_Empty(t *testing.T) {
	validate := validator.New()
	authSvc := new(mockAuthService)
	listingSvc := new(mockListingService)

	h := listings.New(authSvc, listingSvc, validate)

	listingSvc.
		On("List", mock.Anything, mock.Anything).
		Return([]*models.ListingWithAuthor{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ListListings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var out []*models.ListingWithAuthor
	err := json.NewDecoder(resp.Body).Decode(&out)
	require.NoError(t, err)
	require.Empty(t, out)
}

func TestListListings_ErrorFromService(t *testing.T) {
	validate := validator.New()
	authSvc := new(mockAuthService)
	listingSvc := new(mockListingService)

	h := listings.New(authSvc, listingSvc, validate)

	listingSvc.
		On("List", mock.Anything, mock.Anything).
		Return([]*models.ListingWithAuthor{}, context.DeadlineExceeded)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ListListings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestListListings_WithFilters(t *testing.T) {
	validate := validator.New()
	authSvc := new(mockAuthService)
	listingSvc := new(mockListingService)

	h := listings.New(authSvc, listingSvc, validate)

	query := url.Values{}
	query.Set("limit", "5")
	query.Set("offset", "2")
	query.Set("price_min", "100")
	query.Set("price_max", "200")
	query.Set("sort_by", "price")
	query.Set("sort_order", "desc")

	listingSvc.
		On("List", mock.Anything, mock.MatchedBy(func(f storage.ListFilter) bool {
			return f.Limit == 5 &&
				f.Offset == 2 &&
				f.SortBy == "price" &&
				f.SortOrder == "desc" &&
				f.PriceMin != nil && *f.PriceMin == 100 &&
				f.PriceMax != nil && *f.PriceMax == 200
		})).
		Return([]*models.ListingWithAuthor{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)
	w := httptest.NewRecorder()

	h.ListListings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
