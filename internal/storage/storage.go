package storage

import (
	"context"

	"github.com/justcgh9/vk-internship-application/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username, passwordHash string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

type ListingRepository interface {
	CreateListing(ctx context.Context, l *models.Listing) (*models.Listing, error)

	ListListings(ctx context.Context, filter ListFilter) ([]*models.ListingWithAuthor, error)
}

type ListFilter struct {
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
	PriceMin  *float64
	PriceMax  *float64
	ViewerID  *int64
}
