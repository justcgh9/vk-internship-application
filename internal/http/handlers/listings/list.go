package listings

import (
	"net/http"
	"strconv"

	"github.com/justcgh9/vk-internship-application/internal/http/middleware"
	"github.com/justcgh9/vk-internship-application/internal/storage"
	"github.com/justcgh9/vk-internship-application/pkg/httpx"
)

func (h *Handler) ListListings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	filter := storage.ListFilter{
		SortBy:    query.Get("sort_by"),    // "created_at" (default) or "price"
		SortOrder: query.Get("sort_order"), // "asc" or "desc"
		Limit:     parseInt(query.Get("limit"), 10),
		Offset:    parseInt(query.Get("offset"), 0),
	}

	if min := query.Get("price_min"); min != "" {
		if val, err := strconv.ParseFloat(min, 64); err == nil {
			filter.PriceMin = &val
		}
	}
	if max := query.Get("price_max"); max != "" {
		if val, err := strconv.ParseFloat(max, 64); err == nil {
			filter.PriceMax = &val
		}
	}

	if userID, ok := middleware.GetUserID(r.Context()); ok {
		filter.ViewerID = &userID
	}

	listings, err := h.listingSvc.List(r.Context(), filter)
	if err != nil {
		http.Error(w, "failed to fetch listings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, listings)
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return def
}
