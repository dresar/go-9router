package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/go-9router/internal/logging"
	"github.com/dresar/go-9router/internal/providers"
	"github.com/dresar/go-9router/internal/providers/adapters"
	"github.com/dresar/go-9router/internal/storage/repos"
	"github.com/dresar/go-9router/internal/stream"
	"github.com/dresar/go-9router/internal/tokensaver"
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
	if strings.ToLower(r.Header.Get("x-9router-token-saver")) != "off" {
		tokensaver.ApplyTokenSaver(body, settings)
	}

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
	startTime := time.Now()

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
			h.JSONError(w, http.StatusServiceUnavailable, fmt.Sprintf("no active credentials for provider: %s", providerID))
			return
		}
		if sel.AllLocked {
			if lastError == "" {
				lastError = "all accounts temporarily unavailable"
			}
			latencyMs := time.Since(startTime).Milliseconds()
			go h.recordObservability(providerID, modelStr, "", "unavailable", latencyMs, body)
			h.JSONError(w, http.StatusServiceUnavailable, lastError)
			return
		}

		body["model"] = modelID
		providers.LogRequest(providerID, modelID, sel.Credentials.ConnectionName, stream.IsSSERequest(body))

		adapter, known := adapters.GetAdapterWithCreds(providerID, sel.Credentials)
		if !known {
			providers.RecordConnectionEnd(sel.Credentials.ConnectionID)
			h.JSONError(w, http.StatusNotImplemented,
				fmt.Sprintf("provider '%s' not yet implemented. Use Node.js backend for this provider.", providerID),
			)
			return
		}

		req, err := adapter.BuildRequest(r.Context(), body, sel.Credentials)
		if err != nil {
			providers.RecordConnectionEnd(sel.Credentials.ConnectionID)
			h.JSONError(w, http.StatusInternalServerError, "failed to build upstream request")
			return
		}

		result, err := providers.DoUpstreamWithProxy(r.Context(), req, h.DB, sel.Credentials)
		if err != nil {
			latencyMs := time.Since(startTime).Milliseconds()
			providers.RecordConnectionFailure(h.DB, sel.Credentials.ConnectionID, http.StatusBadGateway, err.Error())
			go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, "502", latencyMs, body)
			h.JSONError(w, http.StatusBadGateway, "upstream error")
			return
		}

		if result.Success {
			latencyMs := time.Since(startTime).Milliseconds()
			providers.RecordConnectionSuccess(h.DB, sel.Credentials.ConnectionID, latencyMs)
			go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, "success", latencyMs, body)

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
			latencyMs := time.Since(startTime).Milliseconds()
			logging.Warn("FALLBACK", "account unavailable, trying next",
				"provider", providerID,
				"account", sel.Credentials.ConnectionName,
				"status", result.Status,
			)
			providers.RecordConnectionFailure(h.DB, sel.Credentials.ConnectionID, result.Status, result.Error)
			go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, fmt.Sprintf("%d", result.Status), latencyMs, body)
			exclude[sel.Credentials.ConnectionID] = true
			lastError = result.Error
			if result.Response != nil {
				result.Response.Body.Close()
			}
			continue
		}

		if result.Response != nil {
			latencyMs := time.Since(startTime).Milliseconds()
			providers.RecordConnectionEnd(sel.Credentials.ConnectionID)
			go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, fmt.Sprintf("%d", result.Status), latencyMs, body)
			defer result.Response.Body.Close()
			stream.ProxyJSON(result.Response, w)
		} else {
			latencyMs := time.Since(startTime).Milliseconds()
			providers.RecordConnectionFailure(h.DB, sel.Credentials.ConnectionID, result.Status, result.Error)
			go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, fmt.Sprintf("%d", result.Status), latencyMs, body)
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
	adapter, known := adapters.GetAdapterWithCreds(providerID, sel.Credentials)
	if !known {
		return nil
	}
	req, err := adapter.BuildRequest(r.Context(), body, sel.Credentials)
	if err != nil {
		return nil
	}
	result, err := providers.DoUpstreamWithProxy(r.Context(), req, h.DB, sel.Credentials)
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

func (h *Handler) recordObservability(providerID, modelStr, connectionID, status string, latencyMs int64, body map[string]any) {
	settings, _ := repos.GetSettings(h.DB)
	if !repos.SettingBool(settings, "enableObservability", true) && !h.Cfg.ObservabilityEnabled {
		return
	}

	pTokens := 0
	if msgs, ok := body["messages"].([]any); ok {
		for _, m := range msgs {
			if mm, ok := m.(map[string]any); ok {
				if content, ok := mm["content"].(string); ok {
					pTokens += len(content) / 4
				}
			}
		}
	}
	if pTokens < 10 {
		pTokens = 15
	}
	cTokens := 60

	reqID := fmt.Sprintf("%s-%s-%s", time.Now().Format("20060102-150405"), randomSuffix(6), cleanModelName(modelStr))

	_ = repos.SaveRequestDetail(h.DB, repos.RequestDetail{
		ID:           reqID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339Nano),
		Provider:     providerID,
		Model:        modelStr,
		ConnectionID: connectionID,
		Status:       status,
		Latency: map[string]any{
			"total": latencyMs,
		},
		Tokens: map[string]any{
			"prompt_tokens":     pTokens,
			"completion_tokens": cTokens,
			"total_tokens":      pTokens + cTokens,
		},
	})

	_ = repos.SaveUsage(h.DB, repos.UsageRecord{
		Timestamp:        time.Now().UTC().Format(time.RFC3339Nano),
		Provider:         providerID,
		Model:            modelStr,
		ConnectionID:     connectionID,
		Endpoint:         "/v1/chat/completions",
		PromptTokens:     pTokens,
		CompletionTokens: cTokens,
		Status:           status,
	})
}

func randomSuffix(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

func cleanModelName(m string) string {
	return strings.ReplaceAll(strings.ReplaceAll(m, "/", "-"), ":", "-")
}

