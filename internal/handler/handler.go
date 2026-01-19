package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mitchross/pvc-pulmber/internal/s3"
)

type Handler struct {
	s3Client *s3.Client
	logger   *slog.Logger
}

func New(s3Client *s3.Client, logger *slog.Logger) *Handler {
	return &Handler{
		s3Client: s3Client,
		logger:   logger,
	}
}

func (h *Handler) HandleExists(w http.ResponseWriter, r *http.Request) {
	// Extract namespace and pvc from path
	// Expected path: /exists/{namespace}/{pvc}
	path := r.URL.Path

	var namespace, pvc string
	// Simple path parsing
	if len(path) > 8 { // "/exists/"
		parts := path[8:] // Remove "/exists/"
		var foundSlash bool
		var i int
		for i = 0; i < len(parts); i++ {
			if parts[i] == '/' {
				namespace = parts[:i]
				foundSlash = true
				break
			}
		}
		if foundSlash && i+1 < len(parts) {
			pvc = parts[i+1:]
		}
	}

	if namespace == "" || pvc == "" {
		h.logger.Warn("invalid request path", "path", path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"exists": false,
			"error":  "invalid path format, expected /exists/{namespace}/{pvc}",
		})
		return
	}

	h.logger.Info("checking backup", "namespace", namespace, "pvc", pvc)

	result := h.s3Client.CheckBackupExists(r.Context(), namespace, pvc)

	h.logger.Info("backup check complete",
		"namespace", namespace,
		"pvc", pvc,
		"exists", result.Exists,
		"keyCount", result.KeyCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) HandleReadyz(w http.ResponseWriter, r *http.Request) {
	// Same as healthz for now
	h.HandleHealthz(w, r)
}
