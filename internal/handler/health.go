package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Sahilkumar121/short_url/internal/httpx"
	"github.com/Sahilkumar121/short_url/internal/middleware"
)

type HealthHandler struct {
	Logger *slog.Logger
}

func NewHealthHandler(logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		Logger: logger,
	}
}

func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {

	reqID := middleware.GetRequestID(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": "v1",
	}); err != nil {
		h.Logger.Error("failed to check health", "err", err, "requestID", reqID)
		httpx.HandleError(w, http.StatusInternalServerError, "something is wrong in server", httpx.ErrInternalServer)
		return
	}
}
