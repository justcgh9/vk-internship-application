package listings

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/justcgh9/vk-internship-application/internal/http/middleware"
	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
	"github.com/justcgh9/vk-internship-application/pkg/httpx"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
)

func (h *Handler) ListListings(w http.ResponseWriter, r *http.Request) {
	log := logger.
		FromContext(r.Context()).
		With("component", "handler").
		With("function", "list_listings")

	log.Info("listings request received", slog.String("method", r.Method), slog.String("url", r.RequestURI))

	query := r.URL.Query()

	filter := storage.ListFilter{
		SortBy:    query.Get("sort_by"),
		SortOrder: query.Get("sort_order"),
		Limit:     parseInt(query.Get("limit"), 10),
		Offset:    parseInt(query.Get("offset"), 0),
	}

	if min := query.Get("price_min"); min != "" {
		if val, err := strconv.ParseFloat(min, 64); err == nil {
			filter.PriceMin = &val
		} else {
			log.Warn("invalid price_min", slog.String("value", min), slog.String("err", err.Error()))
		}
	}
	if max := query.Get("price_max"); max != "" {
		if val, err := strconv.ParseFloat(max, 64); err == nil {
			filter.PriceMax = &val
		} else {
			log.Warn("invalid price_max", slog.String("value", max), slog.String("err", err.Error()))
		}
	}

	if userID, ok := middleware.GetUserID(r.Context()); ok {
		filter.ViewerID = &userID
		log = log.With("viewer_id", userID)
	}

	log.Debug("filter applied", slog.Any("filter", filter))

	listings, err := h.listingSvc.List(r.Context(), filter)
	if err != nil {
		log.Error("failed to list listings", slog.String("err", err.Error()))
		http.Error(w, "failed to fetch listings", http.StatusInternalServerError)
		return
	}

	if listings == nil {
		listings = []*models.ListingWithAuthor{}
	}

	log.Info("listings fetched", slog.Int("count", len(listings)))

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
