package listings

import (
	"encoding/json"
	"net/http"

	"github.com/justcgh9/vk-internship-application/internal/http/middleware"
	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/pkg/httpx"
)

type CreateListingRequest struct {
	Title       string  `json:"title" validate:"required,min=3,max=100"`
	Description string  `json:"description" validate:"required,min=10,max=500"`
	ImageURL    string  `json:"image_url" validate:"required,url"`
	Price       float64 `json:"price" validate:"required,gt=0"`
}

func (h *Handler) CreateListing(w http.ResponseWriter, r *http.Request) {
	var req CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "failed to create listing: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, created)
}
