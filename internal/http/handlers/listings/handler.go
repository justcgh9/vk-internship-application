package listings

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/justcgh9/vk-internship-application/internal/http/middleware"
	"github.com/justcgh9/vk-internship-application/internal/service/auth"
	"github.com/justcgh9/vk-internship-application/internal/service/listing"
)

type Handler struct {
	listingSvc listing.Service
	validator  *validator.Validate
}

func New(listingSvc listing.Service, v *validator.Validate) *Handler {
	return &Handler{
		listingSvc: listingSvc,
		validator:  v,
	}
}

func (h *Handler) Routes(authSvc auth.AuthService) chi.Router {
	r := chi.NewRouter()

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authSvc)) // ensures user is in context
		r.Post("/", h.CreateListing)
	})

	// Public route
	r.Get("/", h.ListListings)

	return r
}
