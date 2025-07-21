package listing

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
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
	log := logger.
		FromContext(ctx).
		With("component", "service", "method", "CreateListing")

	if strings.TrimSpace(l.Title) == "" || len(l.Title) > 100 ||
		len(l.Description) > 1000 || l.Price <= 0 || l.UserID == 0 {
		log.Warn("invalid listing data", slog.Any("listing", l))
		return nil, ErrInvalidListing
	}

	created, err := s.listingRepo.CreateListing(ctx, l)
	if err != nil {
		log.Error("failed to create listing", slog.String("err", err.Error()))
		return nil, err
	}

	log.Info("listing created successfully", slog.Int64("listing_id", created.ID))
	return created, nil
}

func (s *service) List(ctx context.Context, filter storage.ListFilter) ([]*models.ListingWithAuthor, error) {
	log := logger.
		FromContext(ctx).
		With("component", "service", "method", "List")

	listings, err := s.listingRepo.ListListings(ctx, filter)
	if err != nil {
		log.Error("failed to fetch listings", slog.String("err", err.Error()), slog.Any("filter", filter))
		return nil, err
	}

	log.Debug("listings fetched", slog.Int("count", len(listings)))
	return listings, nil
}
