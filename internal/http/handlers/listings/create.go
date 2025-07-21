package listings

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/justcgh9/vk-internship-application/internal/http/middleware"
	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/pkg/httpx"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
)

type CreateListingRequest struct {
	Title       string  `json:"title" validate:"required,min=3,max=100"`
	Description string  `json:"description" validate:"required,min=10,max=500"`
	ImageURL    string  `json:"image_url" validate:"required,url"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

const (
	MaxImageSize = 5 * 1024 * 1024
)

func (h *Handler) CreateListing(w http.ResponseWriter, r *http.Request) {
	log := logger.
		FromContext(r.Context()).
		With("component", "handler").
		With("function", "create_listing")

	log.Info("create listing request received", slog.String("method", r.Method), slog.String("url", r.RequestURI))

	var req CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("error decoding request body", slog.String("err", err.Error()))
		http.Error(w, "invalid JSON", http.StatusUnprocessableEntity)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		log.Error("validation failed", slog.String("err", err.Error()))
		http.Error(w, "invalid listing data", http.StatusUnprocessableEntity)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Warn("unauthorized request - no user ID in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := http.Get(req.ImageURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Warn("image URL request failed or non-200", slog.String("err", err.Error()), slog.Int("status", resp.StatusCode))
		http.Error(w, "invalid image URL", http.StatusBadRequest)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	contentType := resp.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		log.Warn("unsupported image format", slog.String("content_type", contentType))
		http.Error(w, "unsupported image format", http.StatusBadRequest)
		return
	}

	if size := resp.ContentLength; size > MaxImageSize {
		log.Warn("image too large", slog.Int64("size", size))
		http.Error(w, "image too large", http.StatusBadRequest)
		return
	}

	listing := &models.Listing{
		Title:       req.Title,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Price:       req.Price,
		UserID:      userID,
	}

	created, err := h.listingSvc.Create(r.Context(), listing)
	if err != nil {
		log.Error("failed to create listing", slog.String("err", err.Error()))
		http.Error(w, "failed to create listing", http.StatusInternalServerError)
		return
	}

	response := &models.ListingWithAuthor{
		ID:          created.ID,
		Title:       created.Title,
		Description: created.Description,
		ImageURL:    created.ImageURL,
		Price:       created.Price,
		IsOwned:     true,
		CreatedAt:   created.CreatedAt,
	}

	user, err := h.authSvc.GetUser(r.Context(), userID)
	if err == nil {
		response.AuthorLogin = user.Username
	} else {
		log.Warn("failed to fetch author info", slog.String("err", err.Error()))
	}

	log.Info("listing created", slog.Int64("listing_id", created.ID), slog.Int64("user_id", userID))

	httpx.WriteJSON(w, http.StatusCreated, response)
}
