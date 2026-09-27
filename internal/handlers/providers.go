package handlers

import (
	"net/http"
	"strings"

	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleProviders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listProviders(w, r)
	case http.MethodPost:
		h.createProvider(w, r)
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleProviderByID(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r, "/api/providers/")
	if id == "" {
		h.JSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	if strings.Contains(id, "/") {
		h.JSONError(w, http.StatusNotFound, "not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getProvider(w, r, id)
	case http.MethodPut:
		h.updateProvider(w, r, id)
	case http.MethodDelete:
		h.deleteProvider(w, r, id)
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) listProviders(w http.ResponseWriter, r *http.Request) {
	conns, err := repos.ListConnections(h.DB, repos.ConnectionFilter{})
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to fetch providers")
		return
	}
	safe := make([]map[string]any, 0, len(conns))
	for _, c := range conns {
		safe = append(safe, safeConnection(c))
	}
	h.JSON(w, http.StatusOK, map[string]any{"connections": safe})
}

func (h *Handler) getProvider(w http.ResponseWriter, r *http.Request, id string) {
	c, err := repos.GetConnection(h.DB, id)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to fetch provider")
		return
	}
	if c == nil {
		h.JSONError(w, http.StatusNotFound, "not found")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{"connection": safeConnection(*c)})
}

func (h *Handler) createProvider(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := h.DecodeJSON(r, &body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	provider, _ := body["provider"].(string)
	if provider == "" {
		h.JSONError(w, http.StatusBadRequest, "provider is required")
		return
	}
	authType, _ := body["authType"].(string)
	if authType == "" {
		authType = "apikey"
	}
	name, _ := body["name"].(string)
	if name == "" {
		name, _ = body["displayName"].(string)
	}
	apiKey, _ := body["apiKey"].(string)

	psd := map[string]any{}
	if v, ok := body["providerSpecificData"].(map[string]any); ok {
		for k, val := range v {
			psd[k] = val
		}
	}

	conn := repos.Connection{
		Provider:             provider,
		AuthType:             authType,
		Name:                 name,
		APIKey:               apiKey,
		IsActive:             true,
		TestStatus:           "unknown",
		Priority:             1,
		ProviderSpecificData: psd,
	}
	created, err := repos.CreateConnection(h.DB, conn)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to create provider")
		return
	}
	h.JSON(w, http.StatusCreated, map[string]any{"connection": safeConnection(*created)})
}

func (h *Handler) updateProvider(w http.ResponseWriter, r *http.Request, id string) {
	var body map[string]any
	if err := h.DecodeJSON(r, &body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	updated, err := repos.UpdateConnection(h.DB, id, body)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to update provider")
		return
	}
	if updated == nil {
		h.JSONError(w, http.StatusNotFound, "not found")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{"connection": safeConnection(*updated)})
}

func (h *Handler) deleteProvider(w http.ResponseWriter, r *http.Request, id string) {
	if err := repos.DeleteConnection(h.DB, id); err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to delete provider")
		return
	}
	h.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func safeConnection(c repos.Connection) map[string]any {
	m := map[string]any{
		"id":        c.ID,
		"provider":  c.Provider,
		"authType":  c.AuthType,
		"name":      c.Name,
		"email":     c.Email,
		"priority":  c.Priority,
		"isActive":  c.IsActive,
		"createdAt": c.CreatedAt,
		"updatedAt": c.UpdatedAt,
		"testStatus": c.TestStatus,
	}
	if c.ProviderSpecificData != nil {
		psd := map[string]any{}
		for k, v := range c.ProviderSpecificData {
			if k != "apiKey" && k != "accessToken" && k != "refreshToken" && k != "idToken" {
				psd[k] = v
			}
		}
		m["providerSpecificData"] = psd
	}
	return m
}
