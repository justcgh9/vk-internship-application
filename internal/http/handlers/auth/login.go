package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

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
	ctx, span := otel.Tracer("vk-intern-app").Start(r.Context(), "auth.login")
	defer span.End()

	log := logger.
		FromContext(ctx).
		With("component", "handler").
		With("function", "login")

	log.Info("login attempt", slog.String("method", r.Method), slog.String("URL", r.RequestURI))

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("error decoding request body", slog.String("err", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid JSON")
		http.Error(w, "invalid JSON", http.StatusUnprocessableEntity)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		log.Error("error validating request body", slog.String("err", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		http.Error(w, "login and password cannot be empty", http.StatusUnprocessableEntity)
		return
	}

	span.SetAttributes(attribute.String("auth.username", req.Username))

	token, err := h.authSvc.Login(ctx, req.Username, req.Password)
	if err != nil {
		log.Error("error authorizing user", slog.String("err", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "login failed")
		http.Error(w, "unauthorized: invalid login or password", http.StatusUnauthorized)
		return
	}

	log.Info("attempt succeeded")
	span.SetStatus(codes.Ok, "login successful")

	httpx.WriteJSON(w, http.StatusOK, LoginResponse{Token: token})
}
