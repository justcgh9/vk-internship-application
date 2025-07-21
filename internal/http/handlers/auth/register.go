package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/pkg/httpx"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
)

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

type RegisterResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	log := logger.
		FromContext(r.Context()).
		With("component", "handler").
		With("function", "register")

	log.Info("register attempt", slog.String("method", r.Method), slog.String("url", r.RequestURI))

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("error decoding request body", slog.String("err", err.Error()))
		http.Error(w, "invalid JSON", http.StatusUnprocessableEntity)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		log.Error("validation failed", slog.String("err", err.Error()))
		http.Error(w, "username and password must meet validation constraints", http.StatusUnprocessableEntity)
		return
	}

	user, token, err := h.authSvc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		log.Error("error registering user", slog.String("err", err.Error()))
		http.Error(w, "could not register user", http.StatusInternalServerError)
		return
	}

	log.Info("register succeeded", slog.Int64("user_id", user.ID))

	httpx.WriteJSON(w, http.StatusCreated, RegisterResponse{
		User:  user,
		Token: token,
	})
}
