package handlers

import (
	"net/http"
	"runtime"
)

const version = "0.1.0"

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) HandleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"version": version,
		"go":      runtime.Version(),
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	})
}

func (h *Handler) HandleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"initialized": true,
		"version":     version,
	})
}

func (h *Handler) HandleLocale(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{"locale": "en"})
}

func (h *Handler) HandleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	go func() {
		_ = r.Context().Err()
	}()
}
