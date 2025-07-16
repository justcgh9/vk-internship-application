package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	authsvc "github.com/justcgh9/vk-internship-application/internal/service/auth"
	authhdl "github.com/justcgh9/vk-internship-application/internal/http/handlers/auth"
)

func NewRouter(authService authsvc.AuthService) *chi.Mux {
	r := chi.NewRouter()

	validate := validator.New()

	authHandler := authhdl.New(authService, validate)
	r.Mount("/auth", authHandler.Routes())

	return r
}
