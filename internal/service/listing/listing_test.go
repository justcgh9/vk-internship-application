package listing_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/service/listing"
	"github.com/justcgh9/vk-internship-application/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) CreateListing(ctx context.Context, l *models.Listing) (*models.Listing, error) {
	args := m.Called(ctx, l)
	return args.Get(0).(*models.Listing), args.Error(1)
}

func (m *mockRepo) ListListings(ctx context.Context, filter storage.ListFilter) ([]*models.ListingWithAuthor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*models.ListingWithAuthor), args.Error(1)
}

// --- Tests ---

func TestCreate_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := listing.New(repo)

	input := &models.Listing{
		Title:       "T-shirt",
		Description: "100% cotton",
		Price:       1200,
		UserID:      1,
	}

	expected := &models.Listing{
		ID:          1,
		Title:       "T-shirt",
		Description: "100% cotton",
		Price:       1200,
		UserID:      1,
		CreatedAt:   time.Now(),
	}

	repo.On("CreateListing", mock.Anything, input).Return(expected, nil)

	ctx := context.Background()
	result, err := svc.Create(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestCreate_InvalidInput(t *testing.T) {
	repo := new(mockRepo)
	svc := listing.New(repo)

	invalidInputs := []*models.Listing{
		{Title: "", Description: "Valid", Price: 1000, UserID: 1},
		{Title: strings.Repeat("a", 101), Description: "Valid", Price: 1000, UserID: 1},
		{Title: "Valid", Description: strings.Repeat("a", 1001), Price: 1000, UserID: 1},
		{Title: "Valid", Description: "Valid", Price: 0, UserID: 1},
		{Title: "Valid", Description: "Valid", Price: 1000, UserID: 0},
	}

	ctx := context.Background()

	for _, in := range invalidInputs {
		_, err := svc.Create(ctx, in)
		assert.ErrorIs(t, err, listing.ErrInvalidListing)
	}
}

func TestCreate_RepoError(t *testing.T) {
	repo := new(mockRepo)
	svc := listing.New(repo)

	input := &models.Listing{
		Title:       "Phone",
		Description: "New phone",
		Price:       10000,
		UserID:      2,
	}

	repo.On("CreateListing", mock.Anything, input).Return((*models.Listing)(nil), errors.New("insert failed"))

	ctx := context.Background()
	_, err := svc.Create(ctx, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
	repo.AssertExpectations(t)
}


func TestList_Success(t *testing.T) {
	repo := new(mockRepo)
	svc := listing.New(repo)

	filter := storage.ListFilter{Limit: 10, Offset: 0}
	expected := []*models.ListingWithAuthor{
		{ID: 1, Title: "Shirt", AuthorLogin: "alice", Price: 1000},
	}

	repo.On("ListListings", mock.Anything, filter).Return(expected, nil)

	ctx := context.Background()
	res, err := svc.List(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	repo.AssertExpectations(t)
}

func TestList_RepoError(t *testing.T) {
	repo := new(mockRepo)
	svc := listing.New(repo)

	filter := storage.ListFilter{
		Limit:  10,
		Offset: 0,
	}

	repo.On("ListListings", mock.Anything, filter).Return(([]*models.ListingWithAuthor)(nil), errors.New("query failed"))

	ctx := context.Background()
	_, err := svc.List(ctx, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
	repo.AssertExpectations(t)
}
