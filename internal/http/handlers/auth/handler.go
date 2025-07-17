package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/justcgh9/vk-internship-application/internal/http/middleware"
	"github.com/justcgh9/vk-internship-application/internal/service/auth"
)

type Handler struct {
	authSvc   auth.AuthService
	validator *validator.Validate
}

func New(authSvc auth.AuthService, v *validator.Validate) *Handler {
	return &Handler{
		authSvc:   authSvc,
		validator: v,
	}
}

func (h *Handler) Routes(authSvc auth.AuthService) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(middleware.OptionalAuthMiddleware(authSvc))
		r.Post("/register", h.Register)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.OptionalAuthMiddleware(authSvc))
		r.Post("/login", h.Login)
	})
	

	return r
}
