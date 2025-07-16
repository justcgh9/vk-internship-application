package auth

import (
	"encoding/json"
	"net/http"

	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/pkg/httpx"
)

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

type RegisterRespone struct {
	User	*models.User	`json:"user"`
	Token	string			`json:"token"` 
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.validator.Struct(req); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, token, err := h.authSvc.Register(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, "failed to register: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, RegisterRespone{
		User: user,
		Token: token,
	})
}
