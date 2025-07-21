package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/justcgh9/vk-internship-application/pkg/httpx"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
)

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	log := logger.
		FromContext(r.Context()).
		With("component", "handler").
		With("function", "login")

	log.Info("login attempt", slog.String("method", r.Method), slog.String("URL", r.RequestURI))

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("error decoding request body", slog.String("err", err.Error()))
		http.Error(w, "invalid JSON", http.StatusUnprocessableEntity)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		log.Error("error validating request body", slog.String("err", err.Error()))
		http.Error(w, "login and password cannot be empty", http.StatusUnprocessableEntity)
		return
	}

	token, err := h.authSvc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		log.Error("error authorizing user", slog.String("err", err.Error()))
		http.Error(w, "unauthorized: invalid login or password", http.StatusUnauthorized)
		return
	}

	log.Info("attempt succeeded")

	httpx.WriteJSON(w, http.StatusOK, LoginResponse{Token: token})
}
