package handlers

import (
	"net/http"

	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleKeys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listKeys(w, r)
	case http.MethodPost:
		h.createKey(w, r)
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleKeyByID(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r, "/api/keys/")
	if id == "" {
		h.JSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	switch r.Method {
	case http.MethodDelete:
		if err := repos.DeleteAPIKey(h.DB, id); err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to delete key")
			return
		}
		h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) listKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := repos.ListAPIKeys(h.DB)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to fetch keys")
		return
	}
	if keys == nil {
		keys = []repos.APIKey{}
	}
	h.JSON(w, http.StatusOK, map[string]any{"keys": keys})
}

func (h *Handler) createKey(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := h.DecodeJSON(r, &body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Name == "" {
		h.JSONError(w, http.StatusBadRequest, "Name is required")
		return
	}
	machineID := h.Cfg.MachineIDSalt
	k, err := repos.CreateAPIKey(h.DB, body.Name, machineID)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to create key")
		return
	}
	h.JSON(w, http.StatusCreated, map[string]any{
		"key":       k.Key,
		"name":      k.Name,
		"id":        k.ID,
		"machineId": k.MachineID,
	})
}
