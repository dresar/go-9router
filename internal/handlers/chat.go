package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dresar/go-9router/internal/combo"
	"github.com/dresar/go-9router/internal/logging"
	"github.com/dresar/go-9router/internal/middleware"
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

	// Port-based security: check if this port or context allows direct no-auth bypass
	noAuthPorts := h.GetNoAuthPorts(settings)
	isNoAuth := middleware.IsDirectNoAuthRequest(r, noAuthPorts)

	if !isNoAuth {
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
	}

	modelStr, _ := body["model"].(string)
	if modelStr == "" {
		h.JSONError(w, http.StatusBadRequest, "Missing model")
		return
	}

	// Ensure max_tokens has a healthy floor if unset or very low (< 4096),
	// so reasoning models don't exhaust the entire token budget before producing visible output text.
	if mt, ok := body["max_tokens"].(float64); !ok || mt < 4096 {
		if mct, ok2 := body["max_completion_tokens"].(float64); !ok2 || mct < 4096 {
			body["max_tokens"] = 8192
		}
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
			if isPortOpen("127.0.0.1", 20127) || h.ensureNodeRunning() {
				h.proxyChatToNode(w, r, body, modelStr)
				return
			}
			h.JSONError(w, http.StatusBadRequest, fmt.Sprintf("cannot resolve provider for model: %s", modelStr))
			return
		}

		sel, err := providers.SelectCredentials(h.DB, providerID, exclude)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "credential selection error")
			return
		}
		if sel == nil {
			if isPortOpen("127.0.0.1", 20127) || h.ensureNodeRunning() {
				h.proxyChatToNode(w, r, body, modelStr)
				return
			}
			h.JSONError(w, http.StatusServiceUnavailable, fmt.Sprintf("no active credentials for provider: %s", providerID))
			return
		}
		if sel.AllLocked {
			if lastError == "" {
				lastError = "all accounts temporarily unavailable"
			}
			if isPortOpen("127.0.0.1", 20127) || h.ensureNodeRunning() {
				h.proxyChatToNode(w, r, body, modelStr)
				return
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
			if isPortOpen("127.0.0.1", 20127) || h.ensureNodeRunning() {
				h.proxyChatToNode(w, r, body, modelStr)
				return
			}
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

			defer result.Response.Body.Close()
			if stream.IsSSERequest(body) {
				// Streaming: record observability with estimated tokens, then proxy
				go h.recordObservability(providerID, modelStr, sel.Credentials.ConnectionID, "success", latencyMs, body)
				if err := stream.ProxySSE(r.Context(), w, result.Response); err != nil {
					logging.Warn("STREAM", "SSE proxy error", "err", err)
				}
			} else {
				// Non-streaming: read body, extract real usage, then forward to client
				realUsage, _ := extractUsageFromResponse(result.Response)
				go h.recordObservabilityWithUsage(providerID, modelStr, sel.Credentials.ConnectionID, "success", latencyMs, body, realUsage)
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
		if isPortOpen("127.0.0.1", 20127) || h.ensureNodeRunning() {
			res, errStr := h.callNodeChat(r.Context(), body)
			if res != nil {
				return res, ""
			}
			if errStr != "" {
				return nil, errStr
			}
		}
		return nil, fmt.Sprintf("cannot resolve provider for model: %s", modelStr)
	}

	exclude := map[string]bool{}
	startTime := time.Now()

	for attempt := 0; attempt < 3; attempt++ {
		sel, err := providers.SelectCredentials(h.DB, providerID, exclude)
		if err != nil || sel == nil || sel.AllLocked {
			if attempt == 0 && (isPortOpen("127.0.0.1", 20127) || h.ensureNodeRunning()) {
				res, _ := h.callNodeChat(r.Context(), body)
				if res != nil {
					return res, ""
				}
			}
			break
		}

		body["model"] = modelID
		adapter, known := adapters.GetAdapterWithCreds(providerID, sel.Credentials)
		if !known {
			providers.RecordConnectionEnd(sel.Credentials.ConnectionID)
			if isPortOpen("127.0.0.1", 20127) || h.ensureNodeRunning() {
				res, errStr := h.callNodeChat(r.Context(), body)
				if res != nil {
					return res, ""
				}
				if errStr != "" {
					return nil, errStr
				}
			}
			break
		}

		req, err := adapter.BuildRequest(r.Context(), body, sel.Credentials)
		if err != nil {
			providers.RecordConnectionEnd(sel.Credentials.ConnectionID)
			break
		}

		result, err := providers.DoUpstreamWithProxy(r.Context(), req, h.DB, sel.Credentials)

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

func (h *Handler) GetNoAuthPorts(settings repos.Settings) []string {
	ports := []string{"20129"}
	if h.Cfg != nil {
		if h.Cfg.DirectPort != "" {
			ports = append(ports, strings.TrimSpace(h.Cfg.DirectPort))
		}
		if h.Cfg.NoAuthPorts != "" {
			for _, p := range strings.Split(h.Cfg.NoAuthPorts, ",") {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					ports = append(ports, trimmed)
				}
			}
		}
	}
	if settings != nil {
		if dp := repos.SettingStr(settings, "directPort", ""); dp != "" {
			ports = append(ports, dp)
		}
		if val, ok := settings["noAuthPorts"]; ok {
			if str, ok := val.(string); ok && str != "" {
				for _, p := range strings.Split(str, ",") {
					if trimmed := strings.TrimSpace(p); trimmed != "" {
						ports = append(ports, trimmed)
					}
				}
			}
		}
	}
	return ports
}

var (
	comboCacheMu sync.RWMutex
	cachedCombos = make(map[string]*cachedComboEntry)
)

type cachedComboEntry struct {
	combo     *repos.Combo
	expiresAt time.Time
}

func resolveCombo(db *sql.DB, modelStr string) *repos.Combo {
	name := strings.TrimPrefix(modelStr, "combo:")
	name = strings.TrimPrefix(name, "combo/")

	comboCacheMu.RLock()
	if c, ok := cachedCombos[name]; ok && time.Now().Before(c.expiresAt) {
		comboCacheMu.RUnlock()
		return c.combo
	}
	comboCacheMu.RUnlock()

	combo, err := repos.GetComboByName(db, name)
	if err == nil && combo != nil {
		comboCacheMu.Lock()
		cachedCombos[name] = &cachedComboEntry{combo: combo, expiresAt: time.Now().Add(5 * time.Second)}
		comboCacheMu.Unlock()
		return combo
	}
	combo, err = repos.GetComboByName(db, modelStr)
	if err == nil && combo != nil {
		comboCacheMu.Lock()
		cachedCombos[name] = &cachedComboEntry{combo: combo, expiresAt: time.Now().Add(5 * time.Second)}
		comboCacheMu.Unlock()
		return combo
	}
	return nil
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

// TokenUsage holds real token counts extracted from provider response.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	CachedTokens     int
	CacheCreation    int
}

// extractUsageFromResponse reads a JSON response body and extracts token usage.
// It returns the extracted usage (may be zeroed if not found) and the original body bytes for re-streaming.
func extractUsageFromResponse(resp *http.Response) (TokenUsage, []byte) {
	if resp == nil || resp.Body == nil {
		return TokenUsage{}, nil
	}
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if err != nil {
		return TokenUsage{}, bodyBytes
	}

	var usage TokenUsage
	// Try standard OpenAI-compatible usage object
	var parsed struct {
		Usage struct {
			PromptTokens            int `json:"prompt_tokens"`
			CompletionTokens        int `json:"completion_tokens"`
			TotalTokens             int `json:"total_tokens"`
			PromptTokensDetails     *struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"prompt_tokens_details"`
			CompletionTokensDetails *struct{} `json:"completion_tokens_details"`
		} `json:"usage"`
		// Anthropic / Claude style
		UsageAnthropic struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		} `json:"usage_metadata"`
		// Gemini style
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			CachedContentTokenCount int `json:"cachedContentTokenCount"`
		} `json:"usageMetadata"`
	}
	if json.Unmarshal(bodyBytes, &parsed) == nil {
		// OpenAI/standard
		if parsed.Usage.PromptTokens > 0 || parsed.Usage.CompletionTokens > 0 {
			usage.PromptTokens = parsed.Usage.PromptTokens
			usage.CompletionTokens = parsed.Usage.CompletionTokens
			if parsed.Usage.PromptTokensDetails != nil {
				usage.CachedTokens = parsed.Usage.PromptTokensDetails.CachedTokens
			}
		}
		// Anthropic usage_metadata fallback
		if usage.PromptTokens == 0 && parsed.UsageAnthropic.InputTokens > 0 {
			usage.PromptTokens = parsed.UsageAnthropic.InputTokens
			usage.CompletionTokens = parsed.UsageAnthropic.OutputTokens
			usage.CachedTokens = parsed.UsageAnthropic.CacheReadInputTokens
			usage.CacheCreation = parsed.UsageAnthropic.CacheCreationInputTokens
		}
		// Gemini usageMetadata fallback
		if usage.PromptTokens == 0 && parsed.UsageMetadata.PromptTokenCount > 0 {
			usage.PromptTokens = parsed.UsageMetadata.PromptTokenCount
			usage.CompletionTokens = parsed.UsageMetadata.CandidatesTokenCount
			usage.CachedTokens = parsed.UsageMetadata.CachedContentTokenCount
		}
	}
	return usage, bodyBytes
}

func (h *Handler) recordObservability(providerID, modelStr, connectionID, status string, latencyMs int64, body map[string]any) {
	h.recordObservabilityWithUsage(providerID, modelStr, connectionID, status, latencyMs, body, TokenUsage{})
}

func (h *Handler) recordObservabilityWithUsage(providerID, modelStr, connectionID, status string, latencyMs int64, body map[string]any, realUsage TokenUsage) {
	settings, _ := repos.GetSettings(h.DB)
	if !repos.SettingBool(settings, "enableObservability", true) && !h.Cfg.ObservabilityEnabled {
		return
	}

	pTokens := realUsage.PromptTokens
	cTokens := realUsage.CompletionTokens
	cachedTokens := realUsage.CachedTokens
	cacheCreation := realUsage.CacheCreation

	// Estimate from request body if response didn't have real token data
	if pTokens == 0 {
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
	}
	if cTokens == 0 {
		cTokens = 60
	}

	reqID := fmt.Sprintf("%s-%s-%s", time.Now().Format("20060102-150405"), randomSuffix(6), cleanModelName(modelStr))

	tokensMap := map[string]any{
		"prompt_tokens":     pTokens,
		"completion_tokens": cTokens,
		"total_tokens":      pTokens + cTokens,
	}
	if cachedTokens > 0 {
		tokensMap["cached_tokens"] = cachedTokens
	}
	if cacheCreation > 0 {
		tokensMap["cache_creation_input_tokens"] = cacheCreation
	}

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
		Tokens: tokensMap,
	})

	tokensJson, _ := json.Marshal(tokensMap)
	_ = repos.SaveUsage(h.DB, repos.UsageRecord{
		Timestamp:        time.Now().UTC().Format(time.RFC3339Nano),
		Provider:         providerID,
		Model:            modelStr,
		ConnectionID:     connectionID,
		Endpoint:         "/v1/chat/completions",
		PromptTokens:     pTokens,
		CompletionTokens: cTokens,
		Status:           status,
		Tokens:           string(tokensJson),
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

func (h *Handler) ensureNodeRunning() bool {
	if isPortOpen("127.0.0.1", 20127) {
		return true
	}
	EnsureNextServer()
	return isPortOpen("127.0.0.1", 20127)
}

func (h *Handler) proxyChatToNode(w http.ResponseWriter, r *http.Request, body map[string]any, modelStr string) {
	bodyCopy := copyMap(body)
	if modelStr != "" {
		bodyCopy["model"] = modelStr
	}
	bodyBytes, err := json.Marshal(bodyCopy)
	if err != nil {
		h.JSONError(w, http.StatusBadRequest, "failed to serialize chat body")
		return
	}

	nodeURL := "http://127.0.0.1:20127/api/v1/chat/completions"
	nodeReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, nodeURL, bytes.NewReader(bodyBytes))
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to create node request")
		return
	}

	nodeReq.Header.Set("Content-Type", "application/json")
	nodeReq.Header.Set("Accept", "text/event-stream, application/json")

	// Forward client Authorization if present, or attach local internal API key
	clientAuth := r.Header.Get("Authorization")
	if clientAuth == "" {
		clientAuth = r.Header.Get("x-api-key")
	}
	if clientAuth != "" {
		if !strings.HasPrefix(strings.ToLower(clientAuth), "bearer ") && !strings.Contains(clientAuth, " ") {
			nodeReq.Header.Set("Authorization", "Bearer "+clientAuth)
		} else {
			nodeReq.Header.Set("Authorization", clientAuth)
		}
	} else {
		if keys, err := repos.ListAPIKeys(h.DB); err == nil {
			for _, k := range keys {
				if k.IsActive {
					nodeReq.Header.Set("Authorization", "Bearer "+k.Key)
					break
				}
			}
		}
	}

	for _, hKey := range []string{"anthropic-version", "anthropic-beta", "x-9router-token-saver"} {
		if v := r.Header.Get(hKey); v != "" {
			nodeReq.Header.Set(hKey, v)
		}
	}

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(nodeReq)
	if err != nil {
		h.JSONError(w, http.StatusBadGateway, fmt.Sprintf("Node.js backend error: %v", err))
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		for _, val := range v {
			w.Header().Add(k, val)
		}
	}
	w.WriteHeader(resp.StatusCode)

	if stream.IsSSERequest(body) || strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		flusher, ok := w.(http.Flusher)
		buf := make([]byte, 4096)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				if ok {
					flusher.Flush()
				}
			}
			if readErr != nil {
				break
			}
		}
	} else {
		io.Copy(w, resp.Body)
	}
}

func (h *Handler) callNodeChat(ctx context.Context, body map[string]any) (*providers.UpstreamResult, string) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, "marshal error"
	}
	nodeURL := "http://127.0.0.1:20127/api/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, nodeURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "request error"
	}
	req.Header.Set("Content-Type", "application/json")
	if stream.IsSSERequest(body) {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}

	if keys, err := repos.ListAPIKeys(h.DB); err == nil {
		for _, k := range keys {
			if k.IsActive {
				req.Header.Set("Authorization", "Bearer "+k.Key)
				break
			}
		}
	}

	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err.Error()
	}
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		return nil, string(respBody)
	}
	return &providers.UpstreamResult{
		Response: resp,
		Status:   resp.StatusCode,
		Success:  true,
	}, ""
}


