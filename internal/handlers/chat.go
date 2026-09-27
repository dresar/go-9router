package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dresar/go-9router/internal/logging"
	"github.com/dresar/go-9router/internal/providers"
	"github.com/dresar/go-9router/internal/providers/adapters"
	"github.com/dresar/go-9router/internal/storage/repos"
	"github.com/dresar/go-9router/internal/stream"
)

func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, _ := repos.GetSettings(h.DB)
	requireKey := repos.SettingBool(settings, "requireApiKey", h.Cfg.RequireAPIKey)
	if requireKey {
		apiKey := extractAPIKey(r)
		if apiKey == "" {
			h.JSONError(w, http.StatusUnauthorized, "Missing API key")
			return
		}
		valid, _ := repos.ValidateAPIKey(h.DB, apiKey)
		if !valid {
			h.JSONError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}
	}

	modelStr, _ := body["model"].(string)
	if modelStr == "" {
		h.JSONError(w, http.StatusBadRequest, "Missing model")
		return
	}

	if comboModels := resolveCombo(h.DB, modelStr); len(comboModels) > 0 {
		h.handleComboChat(w, r, body, comboModels)
		return
	}

	h.handleSingleChat(w, r, body, modelStr)
}

func (h *Handler) handleComboChat(w http.ResponseWriter, r *http.Request, body map[string]any, models []string) {
	for _, m := range models {
		bodyCopy := copyMap(body)
		bodyCopy["model"] = m
		result := h.trySingleChat(r, bodyCopy, m)
		if result == nil {
			continue
		}
		defer result.Response.Body.Close()
		if stream.IsSSERequest(body) {
			stream.ProxySSE(r.Context(), w, result.Response)
		} else {
			stream.ProxyJSON(result.Response, w)
		}
		return
	}
	h.JSONError(w, http.StatusServiceUnavailable, "all combo models unavailable")
}

func (h *Handler) handleSingleChat(w http.ResponseWriter, r *http.Request, body map[string]any, modelStr string) {
	exclude := map[string]bool{}
	var lastError string

	for {
		providerID, modelID, ok := providers.ResolveModelProvider(modelStr, h.DB)
		if !ok {
			h.JSONError(w, http.StatusBadRequest, fmt.Sprintf("cannot resolve provider for model: %s", modelStr))
			return
		}

		sel, err := providers.SelectCredentials(h.DB, providerID, exclude)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "credential selection error")
			return
		}
		if sel == nil {
			h.JSONError(w, http.StatusNotFound, fmt.Sprintf("no active credentials for provider: %s", providerID))
			return
		}
		if sel.AllLocked {
			if lastError == "" {
				lastError = "all accounts temporarily unavailable"
			}
			h.JSONError(w, http.StatusServiceUnavailable, lastError)
			return
		}

		body["model"] = modelID
		providers.LogRequest(providerID, modelID, sel.Credentials.ConnectionName, stream.IsSSERequest(body))

		adapter, known := adapters.GetAdapter(providerID)
		if !known {
			h.JSONError(w, http.StatusNotImplemented,
				fmt.Sprintf("provider '%s' not yet implemented. Use Node.js backend for this provider.", providerID),
			)
			return
		}

		req, err := adapter.BuildRequest(r.Context(), body, sel.Credentials)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to build upstream request")
			return
		}

		result, err := providers.DoUpstream(r.Context(), req)
		if err != nil {
			h.JSONError(w, http.StatusBadGateway, "upstream error")
			return
		}

		if result.Success {
			providers.ClearError(h.DB, sel.Credentials.ConnectionID)
			defer result.Response.Body.Close()
			if stream.IsSSERequest(body) {
				if err := stream.ProxySSE(r.Context(), w, result.Response); err != nil {
					logging.Warn("STREAM", "SSE proxy error", "err", err)
				}
			} else {
				if err := stream.ProxyJSON(result.Response, w); err != nil {
					logging.Warn("CHAT", "JSON proxy error", "err", err)
				}
			}
			return
		}

		if providers.ShouldFallback(result.Status) {
			logging.Warn("FALLBACK", "account unavailable, trying next",
				"provider", providerID,
				"account", sel.Credentials.ConnectionName,
				"status", result.Status,
			)
			providers.MarkUnavailable(h.DB, sel.Credentials.ConnectionID, result.Status)
			exclude[sel.Credentials.ConnectionID] = true
			lastError = result.Error
			if result.Response != nil {
				result.Response.Body.Close()
			}
			continue
		}

		if result.Response != nil {
			defer result.Response.Body.Close()
			stream.ProxyJSON(result.Response, w)
		} else {
			h.JSONError(w, result.Status, result.Error)
		}
		return
	}
}

func (h *Handler) trySingleChat(r *http.Request, body map[string]any, modelStr string) *providers.UpstreamResult {
	providerID, modelID, ok := providers.ResolveModelProvider(modelStr, h.DB)
	if !ok {
		return nil
	}
	sel, err := providers.SelectCredentials(h.DB, providerID, nil)
	if err != nil || sel == nil || sel.AllLocked {
		return nil
	}
	body["model"] = modelID
	adapter, known := adapters.GetAdapter(providerID)
	if !known {
		return nil
	}
	req, err := adapter.BuildRequest(r.Context(), body, sel.Credentials)
	if err != nil {
		return nil
	}
	result, err := providers.DoUpstream(r.Context(), req)
	if err != nil || !result.Success {
		if result != nil && result.Response != nil {
			result.Response.Body.Close()
		}
		return nil
	}
	return result
}

func resolveCombo(db *sql.DB, modelStr string) []string {
	combo, err := repos.GetComboByName(db, modelStr)
	if err != nil || combo == nil {
		return nil
	}
	return combo.Models
}

func extractAPIKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return r.Header.Get("x-api-key")
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
