package listing

import (
	"context"
	"errors"
	"strings"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
)

var (
	ErrInvalidListing = errors.New("invalid listing data")
)

type Service interface {
	Create(ctx context.Context, l *models.Listing) (*models.Listing, error)
	List(ctx context.Context, filter storage.ListFilter) ([]*models.ListingWithAuthor, error)
}

type service struct {
	listingRepo storage.ListingRepository
}

func New(listingRepo storage.ListingRepository) Service {
	return &service{listingRepo: listingRepo}
}

func (s *service) Create(ctx context.Context, l *models.Listing) (*models.Listing, error) {
	if strings.TrimSpace(l.Title) == "" || len(l.Title) > 100 ||
		len(l.Description) > 1000 || l.Price <= 0 || l.UserID == 0 {
		return nil, ErrInvalidListing
	}
	return s.listingRepo.CreateListing(ctx, l)
}

func (s *service) List(ctx context.Context, filter storage.ListFilter) ([]*models.ListingWithAuthor, error) {
	return s.listingRepo.ListListings(ctx, filter)
}
