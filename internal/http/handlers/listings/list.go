package listings

import (
	"log/slog"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/justcgh9/vk-internship-application/internal/http/middleware"
	"github.com/justcgh9/vk-internship-application/internal/models"
	"github.com/justcgh9/vk-internship-application/internal/storage"
	"github.com/justcgh9/vk-internship-application/pkg/httpx"
	"github.com/justcgh9/vk-internship-application/pkg/logger"
)

func (h *Handler) ListListings(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("vk-intern-app").Start(r.Context(), "listings.list")
	defer span.End()

	log := logger.
		FromContext(ctx).
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

	span.SetAttributes(
		attribute.String("listings.sort_by", filter.SortBy),
		attribute.String("listings.sort_order", filter.SortOrder),
		attribute.Int("listings.limit", filter.Limit),
		attribute.Int("listings.offset", filter.Offset),
	)

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

	if userID, ok := middleware.GetUserID(ctx); ok {
		filter.ViewerID = &userID
		log = log.With("viewer_id", userID)
		span.SetAttributes(attribute.Int64("listings.viewer_id", userID))
	}

	log.Debug("filter applied", slog.Any("filter", filter))

	listings, err := h.listingSvc.List(ctx, filter)
	if err != nil {
		log.Error("failed to list listings", slog.String("err", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "listings query failed")
		http.Error(w, "failed to fetch listings", http.StatusInternalServerError)
		return
	}

	span.SetStatus(codes.Ok, "listings fetched")
	span.SetAttributes(attribute.Int("listings.count", len(listings)))

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
