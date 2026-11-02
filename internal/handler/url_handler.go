package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Sahilkumar121/short_url/internal/httpx"
	"github.com/Sahilkumar121/short_url/internal/model"
	"github.com/Sahilkumar121/short_url/internal/service"
	"github.com/jackc/pgx/v5"
)

type UrlHandler struct {
	Logger  *slog.Logger
	Service *service.URLService
}

func NewUrlHandler(logger *slog.Logger, service *service.URLService) *UrlHandler {
	return &UrlHandler{
		Logger:  logger,
		Service: service,
	}
}

// GetShortCode: it response with a short code when a long url is pass
func (h *UrlHandler) GetShortCode(w http.ResponseWriter, r *http.Request) {

	// url request data
	var req model.UrlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Error("failed decode error", "err", err, "code", 400)
		httpx.HandleError(w, http.StatusBadRequest, "please provide long url", httpx.ErrBadRequest)
		return
	}

	var url *model.URL
	url, err := h.Service.CreateShortCode(r.Context(), req.LongUrl)
	if err != nil {
		h.Logger.Error("failed url service", "err", err)
		httpx.HandleError(w, http.StatusInternalServerError, "something went wrong, try again", httpx.ErrInternalServer)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"short_code": url.ShortCode}); err != nil {
		return
	}
}

func (h *UrlHandler) GetLongUrl(w http.ResponseWriter, r *http.Request) {

	req := r.PathValue("short_code")

	url, err := h.Service.GenerateLongUrl(r.Context(), req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.HandleError(w, http.StatusNotFound, "No row found", httpx.ErrNotFound)
			return
		}

		h.Logger.Error("failed url service", "err", err)
		httpx.HandleError(w, http.StatusInternalServerError, "something went wrong, try again", httpx.ErrInternalServer)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"long_url": url.LongUrl}); err != nil {
		return
	}
}
