package auth

import (
	"github.com/go-chi/chi/v5"
	"github.com/justcgh9/vk-internship-application/internal/service/auth"
	"github.com/go-playground/validator/v10"
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

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)

	return r
}
