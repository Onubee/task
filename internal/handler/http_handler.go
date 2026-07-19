package handler

import (
	"net/http"
	"time"

	"github.com/Onubee/task/internal/logger"
	"github.com/Onubee/task/internal/service"
)

type HttpHandler struct {
	command service.CommandService
	query   service.QueryService
	logger  *logger.Logger
}

func NewHttpHandler(
	cmd service.CommandService,
	q service.QueryService,
	logger *logger.Logger,
) *HttpHandler {
	return &HttpHandler{
		command: cmd,
		query:   q,
		logger:  logger,
	}
}

func (h *HttpHandler) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, ErrInvalidMethod)
		return
	}

	ctx := r.Context()

	status, err := h.command.GetStatus()
	if err == nil && status != nil {
		respondWithError(w, http.StatusConflict, service.ErrDownloadInProgress)
		return
	}

	go func() {
		h.logger.Info("🚀 Starting download...")
		if err := h.command.Download(ctx); err != nil {
			h.logger.Error("❌ Download failed: %v", err)
		} else {
			h.logger.Info("✅ Download completed successfully")
			if q, ok := h.query.(interface{ InvalidateCache() }); ok {
				q.InvalidateCache()
			}
		}
	}()

	RespondSuccess(w, map[string]interface{}{
		"status":   "download_started",
		"message":  "Download process started asynchronously",
		"duration": time.Since(startTime).String(),
	})
}

func (h *HttpHandler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, ErrInvalidMethod)
		return
	}

	stats, err := h.query.GetStats(r.Context())
	if err != nil {
		h.logger.Error("❌ Failed to get stats: %v", err)
		respondWithError(w, http.StatusInternalServerError, ErrInternalServer)
		return
	}

	RespondSuccess(w, stats)
}

func (h *HttpHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	health, err := h.query.GetHealth(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	statusCode := http.StatusOK
	if health.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	RespondJSON(w, statusCode, health)
}
