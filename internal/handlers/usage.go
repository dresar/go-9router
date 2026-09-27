package handlers

import (
	"net/http"

	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleUsageStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	stats, err := repos.GetUsageStats(h.DB)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to get usage stats")
		return
	}
	h.JSON(w, http.StatusOK, stats)
}

func (h *Handler) HandleUsageHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	q := r.URL.Query()
	filter := repos.UsageFilter{
		Provider:     q.Get("provider"),
		Model:        q.Get("model"),
		ConnectionID: q.Get("connectionId"),
	}
	records, total, err := repos.ListUsage(h.DB, filter)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to get usage history")
		return
	}
	if records == nil {
		records = []repos.UsageRecord{}
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"history": records,
		"total":   total,
	})
}

func (h *Handler) HandleUsageStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.JSONError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}
	w.Write([]byte("data: {\"type\":\"connected\"}\n\n"))
	flusher.Flush()
	<-r.Context().Done()
}

func (h *Handler) HandleRequestDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	details, err := repos.ListRequestDetails(h.DB, 50, 0)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to get request details")
		return
	}
	if details == nil {
		details = []repos.RequestDetail{}
	}
	h.JSON(w, http.StatusOK, map[string]any{"details": details})
}
