package storage

import (
	"context"

	"github.com/justcgh9/vk-internship-application/internal/models"
)

//go:generate mockery --name=UserRepository
type UserRepository interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

//go:generate mockery --name=ListingRepository
type ListingRepository interface {
	CreateListing(ctx context.Context, l *models.Listing) (*models.Listing, error)

	// ListListings returns paginated listings with optional filters and sorting
	ListListings(ctx context.Context, filter ListFilter) ([]*models.ListingWithAuthor, error)
}

// Filtering and pagination options for listings
type ListFilter struct {
	Limit     int
	Offset    int
	SortBy    string   // "created_at" or "price"
	SortOrder string   // "asc" or "desc"
	PriceMin  *float64 // optional
	PriceMax  *float64 // optional
	ViewerID  *int64   // optional, used to mark user's own listings
}
