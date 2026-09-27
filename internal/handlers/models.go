package handlers

import (
	"net/http"
	"strings"

	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	conns, _ := repos.ListConnections(h.DB, repos.ConnectionFilter{})
	models := buildModelList(conns)
	h.JSON(w, http.StatusOK, map[string]any{"data": models, "object": "list"})
}

func (h *Handler) HandleModelAliases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		aliases, _ := repos.KVList(h.DB, "modelAliases")
		h.JSON(w, http.StatusOK, map[string]any{"aliases": aliases})
	case http.MethodPost:
		var body map[string]any
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		alias, _ := body["alias"].(string)
		target, _ := body["target"].(string)
		if alias == "" || target == "" {
			h.JSONError(w, http.StatusBadRequest, "alias and target are required")
			return
		}
		if err := repos.KVSet(h.DB, "modelAliases", alias, target); err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to set alias")
			return
		}
		h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	case http.MethodDelete:
		var body map[string]any
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		alias, _ := body["alias"].(string)
		if alias == "" {
			h.JSONError(w, http.StatusBadRequest, "alias is required")
			return
		}
		if err := repos.KVDelete(h.DB, "modelAliases", alias); err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to delete alias")
			return
		}
		h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleCustomModels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, _ := repos.KVListSlice(h.DB, "customModels")
		if items == nil {
			items = []any{}
		}
		h.JSON(w, http.StatusOK, map[string]any{"models": items})
	case http.MethodPost:
		var body map[string]any
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		id, _ := body["id"].(string)
		provider, _ := body["providerAlias"].(string)
		mtype, _ := body["type"].(string)
		if mtype == "" {
			mtype = "llm"
		}
		key := strings.Join([]string{provider, id, mtype}, "|")
		if err := repos.KVSet(h.DB, "customModels", key, body); err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to add model")
			return
		}
		h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleV1Models(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	conns, _ := repos.ListConnections(h.DB, repos.ConnectionFilter{})
	models := buildModelList(conns)
	h.JSON(w, http.StatusOK, map[string]any{"data": models, "object": "list"})
}

func buildModelList(conns []repos.Connection) []map[string]any {
	seen := map[string]bool{}
	var out []map[string]any
	for _, c := range conns {
		if !c.IsActive {
			continue
		}
		id := c.Provider + "/default"
		if !seen[id] {
			seen[id] = true
			out = append(out, map[string]any{
				"id":       id,
				"object":   "model",
				"owned_by": c.Provider,
			})
		}
	}
	return out
}
