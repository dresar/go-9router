package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dresar/go-9router/internal/combo"
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

	if c := resolveCombo(h.DB, modelStr); c != nil && len(c.Models) > 0 {
		h.handleComboChat(w, r, body, c)
		return
	}

	h.handleSingleChat(w, r, body, modelStr)
}

func (h *Handler) handleComboChat(w http.ResponseWriter, r *http.Request, body map[string]any, c *repos.Combo) {
	settings, _ := repos.GetSettings(h.DB)
	comboStrategy := "round-robin"
	if s, ok := settings["comboStrategy"].(string); ok && s != "" {
		comboStrategy = s
	}
	if strats, ok := settings["comboStrategies"].(map[string]any); ok {
		if cs, ok := strats[c.Name].(map[string]any); ok {
			if fs, ok := cs["fallbackStrategy"].(string); ok && fs != "" {
				comboStrategy = fs
			}
		}
	}
	stickyLimit := 1
	if sl, ok := settings["comboStickyLimit"].(float64); ok && sl > 0 {
		stickyLimit = int(sl)
	}

	candidateModels := combo.DefaultEngine.GetRotatedModels(c.Name, c.Models, comboStrategy, stickyLimit)
	hasVision := combo.DetectRequiredCapabilities(body)
	if hasVision {
		candidateModels = combo.ReorderByCapabilities(candidateModels, true)
	}

	var lastError string
	for i, m := range candidateModels {
		bodyCopy := copyMap(body)
		bodyCopy["model"] = m
		logging.Info("COMBO", fmt.Sprintf("Trying combo [%s] model %d/%d: %s (strategy: %s)", c.Name, i+1, len(candidateModels), m, comboStrategy))

		result, errStr := h.trySingleChatWithFallback(r, bodyCopy, m)
		if result == nil {
			if errStr != "" {
				lastError = errStr
			}
			continue
		}
		defer result.Response.Body.Close()
		logging.Info("COMBO", fmt.Sprintf("Model %s succeeded for combo %s", m, c.Name))
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

	if lastError == "" {
		lastError = "all combo models unavailable"
	}
	h.JSONError(w, http.StatusServiceUnavailable, lastError)
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

func (h *Handler) trySingleChatWithFallback(r *http.Request, body map[string]any, modelStr string) (*providers.UpstreamResult, string) {
	providerID, modelID, ok := providers.ResolveModelProvider(modelStr, h.DB)
	if !ok {
		return nil, fmt.Sprintf("cannot resolve provider for model: %s", modelStr)
	}

	exclude := map[string]bool{}
	startTime := time.Now()

	for attempt := 0; attempt < 3; attempt++ {
		sel, err := providers.SelectCredentials(h.DB, providerID, exclude)
		if err != nil || sel == nil || sel.AllLocked {
			break
		}

		body["model"] = modelID
		adapter, known := adapters.GetAdapterWithCreds(providerID, sel.Credentials)
		if !known {
			providers.RecordConnectionEnd(sel.Credentials.ConnectionID)
			break
		}

		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		req, err := adapter.BuildRequest(ctx, body, sel.Credentials)
		if err != nil {
			cancel()
			providers.RecordConnectionEnd(sel.Credentials.ConnectionID)
			break
		}

		result, err := providers.DoUpstreamWithProxy(ctx, req, h.DB, sel.Credentials)
		cancel()

		latencyMs := time.Since(startTime).Milliseconds()
		if err != nil {
			providers.RecordConnectionFailure(h.DB, sel.Credentials.ConnectionID, http.StatusBadGateway, err.Error())
			go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, "502", latencyMs, body)
			exclude[sel.Credentials.ConnectionID] = true
			continue
		}

		if result.Success {
			providers.RecordConnectionSuccess(h.DB, sel.Credentials.ConnectionID, latencyMs)
			go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, "success", latencyMs, body)
			return result, ""
		}

		if providers.ShouldFallback(result.Status) {
			providers.RecordConnectionFailure(h.DB, sel.Credentials.ConnectionID, result.Status, result.Error)
			go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, fmt.Sprintf("%d", result.Status), latencyMs, body)
			exclude[sel.Credentials.ConnectionID] = true
			if result.Response != nil {
				result.Response.Body.Close()
			}
			continue
		}

		if result.Response != nil {
			result.Response.Body.Close()
		}
		providers.RecordConnectionEnd(sel.Credentials.ConnectionID)
		return nil, result.Error
	}

	return nil, fmt.Sprintf("provider '%s' temporarily unavailable", providerID)
}

func resolveCombo(db *sql.DB, modelStr string) *repos.Combo {
	combo, err := repos.GetComboByName(db, modelStr)
	if err != nil || combo == nil {
		return nil
	}
	return combo
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

