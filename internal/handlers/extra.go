package handlers

import (
	"net/http"

	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleNodes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		nodes, err := repos.ListNodes(h.DB)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to list nodes")
			return
		}
		if nodes == nil {
			nodes = []repos.ProviderNode{}
		}
		h.JSON(w, http.StatusOK, map[string]any{"nodes": nodes})
	case http.MethodPost:
		var body repos.ProviderNode
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		created, err := repos.CreateNode(h.DB, body)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to create node")
			return
		}
		h.JSON(w, http.StatusCreated, map[string]any{"node": created})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleNodeByID(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r, "/api/provider-nodes/")
	if id == "" {
		h.JSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		n, err := repos.GetNodeByID(h.DB, id)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to get node")
			return
		}
		if n == nil {
			h.JSONError(w, http.StatusNotFound, "not found")
			return
		}
		h.JSON(w, http.StatusOK, map[string]any{"node": n})
	case http.MethodPut:
		var body map[string]any
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		updated, err := repos.UpdateNode(h.DB, id, body)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to update node")
			return
		}
		if updated == nil {
			h.JSONError(w, http.StatusNotFound, "not found")
			return
		}
		h.JSON(w, http.StatusOK, map[string]any{"node": updated})
	case http.MethodDelete:
		if err := repos.DeleteNode(h.DB, id); err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to delete node")
			return
		}
		h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleProxyPools(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pools, err := repos.ListProxyPools(h.DB, false)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to list proxy pools")
			return
		}
		if pools == nil {
			pools = []repos.ProxyPool{}
		}
		h.JSON(w, http.StatusOK, map[string]any{"pools": pools})
	case http.MethodPost:
		var body repos.ProxyPool
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		body.IsActive = true
		created, err := repos.CreateProxyPool(h.DB, body)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to create proxy pool")
			return
		}
		h.JSON(w, http.StatusCreated, map[string]any{"pool": created})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleProxyPoolByID(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r, "/api/proxy-pools/")
	if id == "" {
		h.JSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	switch r.Method {
	case http.MethodGet:
		p, err := repos.GetProxyPoolByID(h.DB, id)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to get proxy pool")
			return
		}
		if p == nil {
			h.JSONError(w, http.StatusNotFound, "not found")
			return
		}
		h.JSON(w, http.StatusOK, map[string]any{"pool": p})
	case http.MethodPut:
		var body map[string]any
		if err := h.DecodeJSON(r, &body); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		updated, err := repos.UpdateProxyPool(h.DB, id, body)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to update proxy pool")
			return
		}
		if updated == nil {
			h.JSONError(w, http.StatusNotFound, "not found")
			return
		}
		h.JSON(w, http.StatusOK, map[string]any{"pool": updated})
	case http.MethodDelete:
		if err := repos.DeleteProxyPool(h.DB, id); err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to delete proxy pool")
			return
		}
		h.JSON(w, http.StatusOK, map[string]bool{"success": true})
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleMessages(w http.ResponseWriter, r *http.Request) {
	h.HandleChat(w, r)
}

func (h *Handler) HandleTags(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"tags": []string{"openai", "anthropic", "gemini", "deepseek", "groq", "xai", "openrouter", "mistral", "local"},
	})
}

func (h *Handler) HandlePricing(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"currency": "USD",
		"models": map[string]any{
			"gpt-4o": map[string]any{"input": 2.50, "output": 10.00},
			"gpt-4o-mini": map[string]any{"input": 0.15, "output": 0.60},
			"claude-3-5-sonnet-20241022": map[string]any{"input": 3.00, "output": 15.00},
			"claude-3-5-haiku-20241022": map[string]any{"input": 0.80, "output": 4.00},
			"deepseek-chat": map[string]any{"input": 0.14, "output": 0.28},
			"deepseek-reasoner": map[string]any{"input": 0.55, "output": 2.19},
		},
	})
}

func (h *Handler) HandleTunnelStatus(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"enabled": false,
		"status":  "stopped",
		"url":     "",
	})
}

func (h *Handler) HandleHeadroomStatus(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"enabled":   false,
		"connected": false,
	})
}

func (h *Handler) HandleCLITools(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"tools": []map[string]any{
			{"id": "claude", "name": "Claude Code", "status": "supported", "env": "ANTHROPIC_BASE_URL"},
			{"id": "cursor", "name": "Cursor", "status": "supported", "env": "OPENAI_BASE_URL"},
			{"id": "aider", "name": "Aider", "status": "supported", "env": "OPENAI_API_BASE"},
			{"id": "opencode", "name": "OpenCode", "status": "supported", "env": "OPENAI_BASE_URL"},
		},
	})
}

func (h *Handler) HandleTranslator(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"active": true,
		"formats": []string{"openai", "anthropic", "gemini", "codex", "kiro"},
	})
}
