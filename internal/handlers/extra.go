package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/dresar/go-9router/internal/providers"
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
		activeOnly := r.URL.Query().Get("isActive") == "true"
		pools, err := repos.ListProxyPools(h.DB, activeOnly)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to list proxy pools")
			return
		}
		if pools == nil {
			pools = []repos.ProxyPool{}
		}
		h.JSON(w, http.StatusOK, map[string]any{
			"proxyPools": pools,
			"pools":      pools,
		})
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
		h.JSON(w, http.StatusCreated, map[string]any{
			"proxyPool": created,
			"pool":      created,
		})
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
	if strings.HasSuffix(id, "/test") {
		h.HandleProxyPoolTest(w, r)
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
		h.JSON(w, http.StatusOK, map[string]any{"proxyPool": p, "pool": p})
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
		h.JSON(w, http.StatusOK, map[string]any{"proxyPool": updated, "pool": updated})

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
		"tunnel": map[string]any{
			"enabled": false,
			"status":  "stopped",
			"url":     "",
		},
		"tailscale": map[string]any{
			"enabled":   false,
			"status":    "stopped",
			"url":       "",
			"installed": false,
		},
		"download": nil,
	})
}

func (h *Handler) HandleTunnelAction(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "tailscale-check") {
		h.JSON(w, http.StatusOK, map[string]any{
			"installed":           false,
			"loggedIn":            false,
			"platform":            "windows",
			"brewAvailable":       false,
			"daemonRunning":       false,
			"customDaemonRunning": false,
			"systemDaemonRunning": false,
			"hasCachedPassword":   false,
		})
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"success": true,
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
	path := r.URL.Path
	if strings.HasSuffix(path, "all-statuses") {
		statuses := map[string]any{
			"claude":       map[string]any{"installed": true, "configured": false, "active": false},
			"codex":        map[string]any{"installed": true, "configured": false, "active": false},
			"opencode":     map[string]any{"installed": true, "configured": false, "active": false},
			"droid":        map[string]any{"installed": false, "configured": false, "active": false},
			"openclaw":     map[string]any{"installed": false, "configured": false, "active": false},
			"hermes":       map[string]any{"installed": false, "configured": false, "active": false},
			"cowork":       map[string]any{"installed": false, "configured": false, "active": false},
			"cline":        map[string]any{"installed": false, "configured": false, "active": false},
			"kilo":         map[string]any{"installed": false, "configured": false, "active": false},
			"deepseek-tui": map[string]any{"installed": false, "configured": false, "active": false},
			"jcode":        map[string]any{"installed": false, "configured": false, "active": false},
			"grok-build":   map[string]any{"installed": false, "configured": false, "active": false},
			"devin":        map[string]any{"installed": false, "configured": false, "active": false},
			"pi":           map[string]any{"installed": false, "configured": false, "active": false},
			"omp":          map[string]any{"installed": false, "configured": false, "active": false},
			"crush":        map[string]any{"installed": false, "configured": false, "active": false},
			"forge":        map[string]any{"installed": false, "configured": false, "active": false},
			"smelt":        map[string]any{"installed": false, "configured": false, "active": false},
			"codewhale":    map[string]any{"installed": false, "configured": false, "active": false},
		}
		h.JSON(w, http.StatusOK, statuses)
		return
	}

	h.JSON(w, http.StatusOK, map[string]any{
		"installed":  true,
		"configured": false,
		"active":     false,
		"tools": []map[string]any{
			{"id": "claude", "name": "Claude Code", "status": "supported", "env": "ANTHROPIC_BASE_URL"},
			{"id": "cursor", "name": "Cursor", "status": "supported", "env": "OPENAI_BASE_URL"},
			{"id": "aider", "name": "Aider", "status": "supported", "env": "OPENAI_API_BASE"},
			{"id": "opencode", "name": "OpenCode", "status": "supported", "env": "OPENAI_BASE_URL"},
		},
	})
}

func (h *Handler) HandleTranslator(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.Contains(path, "console-logs") {
		h.JSON(w, http.StatusOK, map[string]any{"logs": []any{}})
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"active":  true,
		"formats": []string{"openai", "anthropic", "gemini", "codex", "kiro"},
	})
}

func (h *Handler) HandleNodeValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"valid": true,
	})
}

func (h *Handler) HandleProxyPoolTest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/proxy-pools/")
	id := strings.TrimSuffix(trimmed, "/test")
	id = strings.Trim(id, "/")

	if id == "" {
		h.JSONError(w, http.StatusBadRequest, "missing pool id")
		return
	}

	pool, err := repos.GetProxyPoolByID(h.DB, id)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to get proxy pool")
		return
	}
	if pool == nil {
		h.JSONError(w, http.StatusNotFound, "proxy pool not found")
		return
	}

	ok, status, elapsedMs, errStr := providers.TestProxyPool(pool)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	testStatus := "error"
	if ok {
		testStatus = "active"
	}
	var lastErr any
	if !ok {
		lastErr = errStr
	}

	_, _ = repos.UpdateProxyPool(h.DB, id, map[string]any{
		"testStatus":   testStatus,
		"lastTestedAt": now,
		"lastError":    lastErr,
		"isActive":     ok,
	})

	h.JSON(w, http.StatusOK, map[string]any{
		"ok":         ok,
		"status":     status,
		"statusText": http.StatusText(status),
		"error":      lastErr,
		"elapsedMs":  elapsedMs,
		"testedAt":   now,
	})
}


func (h *Handler) HandleProxyPoolDeploy(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"deployUrl": "https://relay.example.com",
	})
}


func (h *Handler) HandlePxpipe(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "health") {
		h.JSON(w, http.StatusOK, map[string]any{"status": "ok", "healthy": true})
		return
	}
	if strings.HasSuffix(path, "stats") {
		h.JSON(w, http.StatusOK, map[string]any{"processed": 0, "savedChars": 0})
		return
	}
	if strings.HasSuffix(path, "logs") {
		h.JSON(w, http.StatusOK, map[string]any{"logs": []any{}})
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"running":     false,
		"installed":   false,
		"enabled":     false,
		"autoInstall": false,
		"minChars":    500,
		"timeoutMs":   5000,
	})
}

func (h *Handler) HandleMediaProviders(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"voices": []any{},
	})
}

func (h *Handler) HandleVersionActions(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "shutdown") {
		h.HandleShutdown(w, r)
		return
	}
	// update
	h.JSON(w, http.StatusOK, map[string]any{
		"currentVersion":  "0.1.0-go",
		"latestVersion":   "0.1.0-go",
		"updateAvailable": false,
	})
}
