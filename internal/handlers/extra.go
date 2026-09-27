package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/go-9router/internal/logging"
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

		conns, _ := repos.ListConnections(h.DB, repos.ConnectionFilter{})
		usageMap := make(map[string]int)
		for _, c := range conns {
			if poolID, ok := c.Data["proxyPoolId"].(string); ok && poolID != "" {
				usageMap[poolID]++
			}
		}

		enriched := make([]map[string]any, len(pools))
		for i, p := range pools {
			m := repos.PoolToData(p)
			m["id"] = p.ID
			m["isActive"] = p.IsActive
			m["testStatus"] = p.TestStatus
			m["createdAt"] = p.CreatedAt
			m["updatedAt"] = p.UpdatedAt
			m["boundConnectionCount"] = usageMap[p.ID]
			enriched[i] = m
		}

		h.JSON(w, http.StatusOK, map[string]any{
			"proxyPools": enriched,
			"pools":      enriched,
		})
	case http.MethodPost:
		var input struct {
			Name        string `json:"name"`
			ProxyURL    string `json:"proxyUrl"`
			NoProxy     string `json:"noProxy"`
			IsActive    *bool  `json:"isActive"`
			StrictProxy bool   `json:"strictProxy"`
			Type        string `json:"type"`
		}
		if err := h.DecodeJSON(r, &input); err != nil {
			h.JSONError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		name := strings.TrimSpace(input.Name)
		proxyURL := strings.TrimSpace(input.ProxyURL)
		if name == "" {
			h.JSONError(w, http.StatusBadRequest, "Name is required")
			return
		}
		if proxyURL == "" {
			h.JSONError(w, http.StatusBadRequest, "Proxy URL is required")
			return
		}
		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}
		pType := input.Type
		if pType == "" {
			pType = "http"
		}

		p := repos.ProxyPool{
			Name:        name,
			ProxyURL:    proxyURL,
			NoProxy:     strings.TrimSpace(input.NoProxy),
			IsActive:    isActive,
			StrictProxy: input.StrictProxy,
			Type:        pType,
		}
		created, err := repos.CreateProxyPool(h.DB, p)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to create proxy pool: "+err.Error())
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
		conns, _ := repos.ListConnections(h.DB, repos.ConnectionFilter{})
		boundCount := 0
		for _, c := range conns {
			if poolID, ok := c.Data["proxyPoolId"].(string); ok && poolID == id {
				boundCount++
			}
		}
		if boundCount > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":                fmt.Sprintf("Cannot delete: %d connection(s) are still using this pool.", boundCount),
				"boundConnectionCount": boundCount,
			})
			return
		}
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
			"gpt-4o":                     map[string]any{"input": 2.50, "output": 10.00},
			"gpt-4o-mini":                map[string]any{"input": 0.15, "output": 0.60},
			"claude-3-5-sonnet-20241022": map[string]any{"input": 3.00, "output": 15.00},
			"claude-3-5-haiku-20241022":  map[string]any{"input": 0.80, "output": 4.00},
			"deepseek-chat":              map[string]any{"input": 0.14, "output": 0.28},
			"deepseek-reasoner":          map[string]any{"input": 0.55, "output": 2.19},
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
		if strings.HasSuffix(path, "/stream") {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache, no-transform")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("X-Accel-Buffering", "no")

			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "streaming unsupported", http.StatusInternalServerError)
				return
			}

			// 1. Send all buffered logs immediately on connect
			buffered := logging.GetBufferedLogs()
			if len(buffered) > 0 {
				initBytes, _ := json.Marshal(map[string]any{
					"type": "init",
					"logs": buffered,
				})
				fmt.Fprintf(w, "data: %s\n\n", initBytes)
				flusher.Flush()
			}

			// 2. Subscribe to new log lines
			logChan := make(chan string, 100)
			logging.AddListener(logChan)
			defer logging.RemoveListener(logChan)

			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()

			ctx := r.Context()
			for {
				select {
				case <-ctx.Done():
					return
				case line, ok := <-logChan:
					if !ok {
						return
					}
					lineBytes, _ := json.Marshal(map[string]any{
						"type": "line",
						"line": line,
					})
					fmt.Fprintf(w, "data: %s\n\n", lineBytes)
					flusher.Flush()
				case <-ticker.C:
					fmt.Fprintf(w, ": keepalive\n\n")
					flusher.Flush()
				}
			}
		}

		if r.Method == http.MethodDelete {
			logging.ClearLogs()
			h.JSON(w, http.StatusOK, map[string]bool{"success": true})
			return
		}

		// GET /api/translator/console-logs
		h.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"logs":    logging.GetBufferedLogs(),
		})
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
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body map[string]string
	if err := h.DecodeJSON(r, &body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	path := r.URL.Path
	poolType := "cloudflare"
	deployURL := ""
	name := ""

	if strings.Contains(path, "cloudflare") {
		poolType = "cloudflare"
		name = strings.TrimSpace(body["projectName"])
		if name == "" {
			name = fmt.Sprintf("cloudflare-relay-%d", time.Now().Unix())
		}
		deployURL = fmt.Sprintf("https://%s.workers.dev", name)
	} else if strings.Contains(path, "vercel") {
		poolType = "vercel"
		name = strings.TrimSpace(body["projectName"])
		if name == "" {
			name = fmt.Sprintf("vercel-relay-%d", time.Now().Unix())
		}
		deployURL = fmt.Sprintf("https://%s.vercel.app", name)
	} else if strings.Contains(path, "deno") {
		poolType = "deno"
		name = strings.TrimSpace(body["projectName"])
		org := strings.TrimSpace(body["orgDomain"])
		if name == "" {
			name = fmt.Sprintf("deno-relay-%d", time.Now().Unix())
		}
		if org != "" {
			deployURL = fmt.Sprintf("https://%s.%s", name, strings.TrimPrefix(org, "."))
		} else {
			deployURL = fmt.Sprintf("https://%s.deno.net", name)
		}
	} else {
		name = "custom-relay"
		deployURL = "https://relay.custom.net"
	}

	created, err := repos.CreateProxyPool(h.DB, repos.ProxyPool{
		Name:        name,
		ProxyURL:    deployURL,
		Type:        poolType,
		IsActive:    true,
		StrictProxy: false,
	})
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to create deployed proxy pool: "+err.Error())
		return
	}

	h.JSON(w, http.StatusCreated, map[string]any{
		"success":   true,
		"deployUrl": deployURL,
		"proxyPool": created,
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
	voices, languages, byLang, _ := providers.GetEdgeTtsVoices()
	h.JSON(w, http.StatusOK, map[string]any{
		"voices":    voices,
		"languages": languages,
		"byLang":    byLang,
	})
}

func (h *Handler) HandleVersionActions(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "shutdown") {
		h.HandleShutdown(w, r)
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"currentVersion":  "0.1.0-go",
		"latestVersion":   "0.1.0-go",
		"updateAvailable": false,
	})
}
