package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
)

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Storage struct {
	db DB
}

func NewStorage(db DB) *Storage {
	return &Storage{db: db}
}

// --- UserRepository ---

func (s *Storage) CreateUser(ctx context.Context, username, passwordHash string) (*models.User, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, created_at
	`, username, passwordHash)

	u := &models.User{}
	err := row.Scan(&u.ID, &u.Username, &u.CreatedAt)
	return u, err
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1
	`, username)

	u := &models.User{}
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (s *Storage) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE id = $1
	`, id)

	u := &models.User{}
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

// --- ListingRepository ---

func (s *Storage) CreateListing(ctx context.Context, l *models.Listing) (*models.Listing, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO listings (title, description, image_url, price, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, l.Title, l.Description, l.ImageURL, l.Price, l.UserID)

	err := row.Scan(&l.ID, &l.CreatedAt)
	return l, err
}

func (s *Storage) ListListings(ctx context.Context, filter storage.ListFilter) ([]*models.ListingWithAuthor, error) {
	query := `
		SELECT 
			l.id, l.title, l.description, l.image_url, l.price, u.username, l.user_id, l.created_at
		FROM listings l
		JOIN users u ON l.user_id = u.id
		WHERE 1=1
	`

	args := []any{}
	argID := 1

	// Add price filtering
	if filter.PriceMin != nil {
		query += fmt.Sprintf(" AND l.price >= $%d", argID)
		args = append(args, *filter.PriceMin)
		argID++
	}
	if filter.PriceMax != nil {
		query += fmt.Sprintf(" AND l.price <= $%d", argID)
		args = append(args, *filter.PriceMax)
		argID++
	}

	// Sorting
	sortBy := "l.created_at"
	if filter.SortBy == "price" {
		sortBy = "l.price"
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []*models.ListingWithAuthor
	for rows.Next() {
		var l models.ListingWithAuthor
		var authorID int64
		if err := rows.Scan(
			&l.ID, &l.Title, &l.Description, &l.ImageURL, &l.Price,
			&l.AuthorLogin, &authorID, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		if filter.ViewerID != nil && *filter.ViewerID == authorID {
			l.IsOwned = true
		}
		listings = append(listings, &l)
	}
	return listings, nil
}
