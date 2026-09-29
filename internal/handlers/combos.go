package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dresar/go-9router/internal/combo"
	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleCombos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		combos, err := repos.ListCombos(h.DB)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to list combos")
			return
		}
		if combos == nil {
			combos = []repos.Combo{}
		}
		h.JSON(w, http.StatusOK, map[string]any{"combos": combos})
	case http.MethodPost:
		var body repos.Combo
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if body.Name == "" {
			h.JSONError(w, http.StatusBadRequest, "name is required")
			return
		}
		if len(body.Models) > 3 {
			body.Models = body.Models[:3]
		}
		created, err := repos.CreateCombo(h.DB, body)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to create combo")
			return
		}
		h.JSON(w, http.StatusCreated, map[string]any{"combo": created})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleComboByID(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r, "/api/combos/")
	if id == "" {
		h.JSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		c, err := repos.GetComboByID(h.DB, id)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to get combo")
			return
		}
		if c == nil {
			h.JSONError(w, http.StatusNotFound, "not found")
			return
		}
		h.JSON(w, http.StatusOK, map[string]any{"combo": c})
	case http.MethodPut:
		var body map[string]any
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		if models, ok := body["models"].([]any); ok {
			strs := make([]string, 0, len(models))
			for _, m := range models {
				if s, ok := m.(string); ok {
					strs = append(strs, s)
				}
			}
			if len(strs) > 3 {
				strs = strs[:3]
			}
			body["models"] = strs
		}
		updated, err := repos.UpdateCombo(h.DB, id, body)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to update combo")
			return
		}
		if updated == nil {
			h.JSONError(w, http.StatusNotFound, "not found")
			return
		}
		h.JSON(w, http.StatusOK, map[string]any{"combo": updated})
	case http.MethodDelete:
		if err := repos.DeleteCombo(h.DB, id); err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to delete combo")
			return
		}
		h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleComboPresets(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	if source == "" {
		source = "claude"
	}
	switch r.Method {
	case http.MethodGet:
		combos, _ := repos.ListCombos(h.DB)
		var existingNames []string
		for _, c := range combos {
			existingNames = append(existingNames, c.Name)
		}
		items := combo.BuildPresets(source, existingNames)
		toCreate := 0
		toSkip := 0
		for _, it := range items {
			if it.Exists {
				toSkip++
			} else {
				toCreate++
			}
		}
		h.JSON(w, http.StatusOK, map[string]any{
			"source":   source,
			"items":    items,
			"toCreate": toCreate,
			"toSkip":   toSkip,
		})
	case http.MethodPost:
		var req struct {
			Source string `json:"source"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Source != "" {
			source = req.Source
		}
		combos, _ := repos.ListCombos(h.DB)
		var existingNames []string
		for _, c := range combos {
			existingNames = append(existingNames, c.Name)
		}
		items := combo.BuildPresets(source, existingNames)
		var created []repos.Combo
		var skipped []string
		for _, it := range items {
			if it.Exists {
				skipped = append(skipped, it.Name)
				continue
			}
			newCombo, err := repos.CreateCombo(h.DB, repos.Combo{
				Name:   it.Name,
				Models: it.Models,
			})
			if err == nil && newCombo != nil {
				created = append(created, *newCombo)
			}
		}
		h.JSON(w, http.StatusOK, map[string]any{
			"source":       source,
			"created":      created,
			"skipped":      skipped,
			"createdCount": len(created),
			"skippedCount": len(skipped),
		})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
