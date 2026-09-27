package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/dresar/go-9router/internal/config"
)

type Handler struct {
	DB  *sql.DB
	Cfg *config.Config
}

func (h *Handler) JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) JSONError(w http.ResponseWriter, status int, msg string) {
	h.JSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) DecodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func MethodRouter(handlers map[string]http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn, ok := handlers[r.Method]
		if !ok {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		fn(w, r)
	}
}

func pathSegment(r *http.Request, after string) string {
	path := r.URL.Path
	idx := len(after)
	if idx >= len(path) {
		return ""
	}
	seg := path[idx:]
	if len(seg) > 0 && seg[0] == '/' {
		seg = seg[1:]
	}
	return seg
}
