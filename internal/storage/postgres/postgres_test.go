package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/jackc/pgx/v5"
	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
	"github.com/justcgh9/vk-internship-application/internal/storage/postgres"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser_Success(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)
	

	rows := pgxmock.NewRows([]string{"id", "username", "created_at"}).
		AddRow(int64(1), "alice", time.Now())

	mockConn.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice", "hashedpassword").
		WillReturnRows(rows)

	ctx := context.Background()
	user, err := store.CreateUser(ctx, "alice", "hashedpassword")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "alice", user.Username)
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestCreateUser_QueryError(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)
	

	mockConn.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice", "hashedpassword").
		WillReturnError(errors.New("db error"))

	ctx := context.Background()
	_, err = store.CreateUser(ctx, "alice", "hashedpassword")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestGetUserByUsername_Success(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)
	

	rows := pgxmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
		AddRow(int64(1), "bob", "hashed", time.Now())

	mockConn.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users WHERE username = \$1`).
		WithArgs("bob").
		WillReturnRows(rows)

	ctx := context.Background()
	user, err := store.GetUserByUsername(ctx, "bob")
	assert.NoError(t, err)
	assert.Equal(t, "bob", user.Username)
	assert.Equal(t, "hashed", user.PasswordHash)
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)
	

	mockConn.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users WHERE username = \$1`).
		WithArgs("nonexistent").
		WillReturnError(pgx.ErrNoRows)

	ctx := context.Background()
	_, err = store.GetUserByUsername(ctx, "nonexistent")
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func setFieldValue(target any, fieldName string, value any) {
	rv := reflect.ValueOf(target)
	for rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	if !rv.CanAddr() {
		panic("target must be addressable")
	}
	if rv.Kind() != reflect.Struct {
		panic("target must be a struct")
	}
	rf := rv.FieldByName(fieldName)
	reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

func TestGetUserByID_Success(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)

	expectedTime := time.Now()
	rows := pgxmock.NewRows([]string{"id", "username", "password_hash", "created_at"}).
		AddRow(int64(2), "alice", "hash123", expectedTime)

	mockConn.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users WHERE id = \$1`).
		WithArgs(int64(2)).
		WillReturnRows(rows)

	ctx := context.Background()
	u, err := store.GetUserByID(ctx, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), u.ID)
	assert.Equal(t, "alice", u.Username)
	assert.Equal(t, "hash123", u.PasswordHash)
	assert.WithinDuration(t, expectedTime, u.CreatedAt, time.Second)
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestGetUserByID_NotFound(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)

	mockConn.ExpectQuery(`SELECT id, username, password_hash, created_at FROM users WHERE id = \$1`).
		WithArgs(int64(99)).
		WillReturnError(pgx.ErrNoRows)

	ctx := context.Background()
	_, err = store.GetUserByID(ctx, 99)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestCreateListing_Success(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)

	listing := &models.Listing{
		Title:       "Cool Shirt",
		Description: "Black shirt with logo",
		ImageURL:    "https://img.com/shirt.png",
		Price:       2500,
		UserID:      1,
	}

	expectedCreatedAt := time.Now()
	rows := pgxmock.NewRows([]string{"id", "created_at"}).AddRow(int64(10), expectedCreatedAt)

	mockConn.ExpectQuery(`INSERT INTO listings`).
		WithArgs(listing.Title, listing.Description, listing.ImageURL, listing.Price, listing.UserID).
		WillReturnRows(rows)

	ctx := context.Background()
	res, err := store.CreateListing(ctx, listing)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), res.ID)
	assert.WithinDuration(t, expectedCreatedAt, res.CreatedAt, time.Second)
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestCreateListing_DBError(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)

	listing := &models.Listing{
		Title:       "Cool Shirt",
		Description: "Black shirt with logo",
		ImageURL:    "https://img.com/shirt.png",
		Price:       2500,
		UserID:      1,
	}

	mockConn.ExpectQuery(`INSERT INTO listings`).
		WithArgs(listing.Title, listing.Description, listing.ImageURL, listing.Price, listing.UserID).
		WillReturnError(errors.New("insert failed"))

	ctx := context.Background()
	_, err = store.CreateListing(ctx, listing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestListListings_Success(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)

	min := 1000.0
	max := 5000.0
	viewerID := int64(1)
	filter := storage.ListFilter{
		PriceMin:  &min,
		PriceMax:  &max,
		SortBy:    "price",
		SortOrder: "asc",
		Limit:     10,
		Offset:    0,
		ViewerID:  &viewerID,
	}

	createdAt := time.Now()
	rows := pgxmock.NewRows([]string{
		"id", "title", "description", "image_url", "price", "username", "user_id", "created_at",
	}).AddRow(1, "Item 1", "desc", "img", 3000, "bob", 1, createdAt)

	mockConn.ExpectQuery(`SELECT .* FROM listings l JOIN users u ON l.user_id = u.id`).
		WithArgs(min, max, filter.Limit, filter.Offset).
		WillReturnRows(rows)

	ctx := context.Background()
	results, err := store.ListListings(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Item 1", results[0].Title)
	assert.True(t, results[0].IsOwned)
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestListListings_DBError(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)

	filter := storage.ListFilter{
		Limit:  5,
		Offset: 0,
	}

	mockConn.ExpectQuery(`SELECT .* FROM listings l JOIN users u ON l.user_id = u.id`).
		WithArgs(filter.Limit, filter.Offset).
		WillReturnError(errors.New("query fail"))

	ctx := context.Background()
	_, err = store.ListListings(ctx, filter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query fail")
	assert.NoError(t, mockConn.ExpectationsWereMet())
}

func TestListListings_ScanError(t *testing.T) {
	mockConn, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockConn.Close()

	store := &postgres.Storage{}
	setFieldValue(store, "db", mockConn)

	filter := storage.ListFilter{
		Limit:  5,
		Offset: 0,
	}

	rows := pgxmock.NewRows([]string{
		"id", "title", "description", "image_url", "price", "username", "user_id", "created_at",
	}).AddRow("not-an-int", "Item", "desc", "img", 1000, "bob", 1, time.Now())

	mockConn.ExpectQuery(`SELECT .* FROM listings l JOIN users u ON l.user_id = u.id`).
		WithArgs(filter.Limit, filter.Offset).
		WillReturnRows(rows)

	ctx := context.Background()
	_, err = store.ListListings(ctx, filter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "destination kind 'int64' not supported for value kind 'string' of column 'id'")
	assert.NoError(t, mockConn.ExpectationsWereMet())
}