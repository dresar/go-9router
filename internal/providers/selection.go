package providers

import (
	"database/sql"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dresar/go-9router/internal/storage/repos"
)

type Credentials struct {
	ConnectionID         string
	ConnectionName       string
	AuthType             string
	APIKey               string
	AccessToken          string
	RefreshToken         string
	IDToken              string
	ExpiresAt            string
	ProjectID            string
	BaseURL              string
	ProxyPoolID          string
	ProviderSpecificData map[string]any
}

type SelectionResult struct {
	Credentials *Credentials
	AllLocked   bool
	RetryAfter  string
}

// In-memory atomic state & metrics for sub-microsecond routing & zero-lock concurrency
var (
	providerCounters  sync.Map // map[string]*atomic.Uint64
	connectionLatency sync.Map // map[string]float64 (rolling exponential moving average)
	inFlightRequests  sync.Map // map[string]*atomic.Int64
	connectionLockMap sync.Map // map[string]time.Time (circuit breaker expiration)
	connectionErrCode sync.Map // map[string]int

	cacheMu   sync.RWMutex
	connCache = make(map[string]cachedConns)
	cacheTTL  = 2 * time.Second
)

type cachedConns struct {
	conns     []repos.Connection
	expiresAt time.Time
}

func InvalidateConnectionCache() {
	cacheMu.Lock()
	connCache = make(map[string]cachedConns)
	cacheMu.Unlock()
}

func getProviderConns(db *sql.DB, provider string) ([]repos.Connection, error) {
	cacheMu.RLock()
	c, ok := connCache[provider]
	cacheMu.RUnlock()
	if ok && time.Now().Before(c.expiresAt) {
		return c.conns, nil
	}

	active := true
	conns, err := repos.ListConnections(db, repos.ConnectionFilter{
		Provider: &provider,
		IsActive: &active,
	})
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	connCache[provider] = cachedConns{
		conns:     conns,
		expiresAt: time.Now().Add(cacheTTL),
	}
	cacheMu.Unlock()
	return conns, nil
}

func SelectCredentials(db *sql.DB, provider string, exclude map[string]bool) (*SelectionResult, error) {
	conns, err := getProviderConns(db, provider)
	if err != nil {
		return nil, err
	}

	if len(conns) == 0 {
		return nil, nil
	}

	cooldownSec := getCooldownSeconds(db)
	available := make([]repos.Connection, 0, len(conns))
	for _, c := range conns {
		if exclude != nil && exclude[c.ID] {
			continue
		}
		if isModelLocked(c, cooldownSec) {
			continue
		}
		available = append(available, c)
	}

	if len(available) == 0 {
		return &SelectionResult{AllLocked: true}, nil
	}

	conn := applyStrategy(db, provider, available)
	RecordConnectionStart(conn.ID)

	return &SelectionResult{
		Credentials: connectionToCredentials(conn),
	}, nil
}

func applyStrategy(db *sql.DB, provider string, available []repos.Connection) repos.Connection {
	if len(available) == 1 {
		return available[0]
	}

	strategy := resolveStrategy(db, provider)

	switch strategy {
	case "fill-first":
		return available[0]

	case "lowest-latency", "least-latency", "fastest":
		var best repos.Connection = available[0]
		bestLat := float64(1000000)
		for _, c := range available {
			if latVal, ok := connectionLatency.Load(c.ID); ok {
				lat := latVal.(float64)
				if lat < bestLat {
					bestLat = lat
					best = c
				}
			} else {
				return c
			}
		}
		return best

	case "least-used", "least-busy":
		var best repos.Connection = available[0]
		minInFlight := int64(1000000)
		for _, c := range available {
			if cntVal, ok := inFlightRequests.Load(c.ID); ok {
				cnt := cntVal.(*atomic.Int64).Load()
				if cnt < minInFlight {
					minInFlight = cnt
					best = c
				}
			} else {
				return c
			}
		}
		return best

	case "round-robin":
		fallthrough
	default:
		val, _ := providerCounters.LoadOrStore(provider, new(atomic.Uint64))
		counter := val.(*atomic.Uint64)
		idx := counter.Add(1) - 1
		return available[idx%uint64(len(available))]
	}
}

func resolveStrategy(db *sql.DB, provider string) string {
	settings, err := repos.GetSettings(db)
	if err != nil || settings == nil {
		return "round-robin"
	}

	if ps, ok := settings["providerStrategies"].(map[string]any); ok {
		if provObj, ok := ps[provider].(map[string]any); ok {
			if strat, ok := provObj["fallbackStrategy"].(string); ok && strat != "" {
				return strings.ToLower(strat)
			}
		}
	}

	if strat, ok := settings["fallbackStrategy"].(string); ok && strat != "" {
		return strings.ToLower(strat)
	}

	return "round-robin"
}

func getCooldownSeconds(db *sql.DB) int {
	settings, err := repos.GetSettings(db)
	if err != nil || settings == nil {
		return 60
	}
	if cd, ok := settings["cooldownDuration"].(float64); ok && cd > 0 {
		return int(cd)
	}
	if cd, ok := settings["cooldownSeconds"].(float64); ok && cd > 0 {
		return int(cd)
	}
	return 60
}

func isModelLocked(c repos.Connection, cooldownSec int) bool {
	if lockUntilVal, ok := connectionLockMap.Load(c.ID); ok {
		if until, ok := lockUntilVal.(time.Time); ok {
			if time.Now().Before(until) {
				return true
			}
			connectionLockMap.Delete(c.ID)
		}
	}

	if c.TestStatus != "unavailable" {
		return false
	}
	if c.LastErrorAt == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339Nano, c.LastErrorAt)
	if err != nil {
		return false
	}
	dur := time.Duration(cooldownSec) * time.Second
	if dur <= 0 {
		dur = 60 * time.Second
	}
	return time.Since(t) < dur
}

func RecordConnectionStart(connectionID string) {
	val, _ := inFlightRequests.LoadOrStore(connectionID, new(atomic.Int64))
	val.(*atomic.Int64).Add(1)
}

func RecordConnectionEnd(connectionID string) {
	if val, ok := inFlightRequests.Load(connectionID); ok {
		val.(*atomic.Int64).Add(-1)
	}
}

func RecordConnectionSuccess(db *sql.DB, connectionID string, latencyMs int64) {
	RecordConnectionEnd(connectionID)

	if latencyMs > 0 {
		val, ok := connectionLatency.Load(connectionID)
		if ok {
			oldAvg := val.(float64)
			newAvg := oldAvg*0.7 + float64(latencyMs)*0.3
			connectionLatency.Store(connectionID, newAvg)
		} else {
			connectionLatency.Store(connectionID, float64(latencyMs))
		}
	}

	connectionLockMap.Delete(connectionID)
	connectionErrCode.Delete(connectionID)

	go ClearError(db, connectionID)
}

func RecordConnectionFailure(db *sql.DB, connectionID string, status int, errMsg string) {
	RecordConnectionEnd(connectionID)

	cooldown := getCooldownSeconds(db)
	if status >= 500 && cooldown > 30 {
		cooldown = 30
	}

	lockUntil := time.Now().Add(time.Duration(cooldown) * time.Second)
	connectionLockMap.Store(connectionID, lockUntil)
	connectionErrCode.Store(connectionID, status)

	go MarkUnavailable(db, connectionID, status)
}

func connectionToCredentials(c repos.Connection) *Credentials {
	name := c.DisplayName
	if name == "" {
		name = c.Name
	}
	if name == "" {
		name = c.Email
	}
	if name == "" {
		name = c.ID[:8]
	}
	baseURL, _ := c.ProviderSpecificData["baseUrl"].(string)
	if baseURL == "" {
		baseURL, _ = c.ProviderSpecificData["baseURL"].(string)
	}
	return &Credentials{
		ConnectionID:         c.ID,
		ConnectionName:       name,
		AuthType:             c.AuthType,
		APIKey:               c.APIKey,
		AccessToken:          c.AccessToken,
		RefreshToken:         c.RefreshToken,
		IDToken:              c.IDToken,
		ExpiresAt:            c.ExpiresAt,
		ProjectID:            c.ProjectID,
		BaseURL:              baseURL,
		ProxyPoolID:          c.ProxyPoolID,
		ProviderSpecificData: c.ProviderSpecificData,
	}
}

func MarkUnavailable(db *sql.DB, connectionID string, status int) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := repos.UpdateConnection(db, connectionID, map[string]any{
		"testStatus":  "unavailable",
		"errorCode":   status,
		"lastErrorAt": now,
	})
	InvalidateConnectionCache()
	return err
}

func ClearError(db *sql.DB, connectionID string) error {
	_, err := repos.UpdateConnection(db, connectionID, map[string]any{
		"testStatus":  "active",
		"lastError":   nil,
		"errorCode":   nil,
		"lastErrorAt": nil,
	})
	InvalidateConnectionCache()
	return err
}

func normalizeProviderAlias(alias string) string {
	lower := strings.ToLower(strings.TrimSpace(alias))
	switch lower {
	case "cf", "cloudflare":
		return "cloudflare-ai"
	case "qd":
		return "qoder"
	case "google":
		return "gemini"
	case "claude":
		return "anthropic"
	case "mimo", "xiaomi":
		return "xiaomi-mimo"
	case "grok":
		return "xai"
	case "copilot":
		return "github"
	case "codebuddy":
		return "codebuddy-intl"
	case "moonshot":
		return "kimi"
	case "cb":
		return "cerebras"
	case "ch":
		return "chutes"
	case "gk":
		return "geraikita"
	case "af", "airforce":
		return "api-airforce"
	case "kgw", "kilogateway":
		return "kilo-gateway"
	default:
		return alias
	}
}

func ResolveModelProvider(modelStr string, db *sql.DB) (provider, model string, ok bool) {
	if modelStr == "" {
		return "", "", false
	}

	// 1. Direct Cloudflare AI model detection (@cf/...)
	if strings.HasPrefix(modelStr, "@cf/") {
		return "cloudflare-ai", modelStr, true
	}

	// 2. Explicit provider/model or alias/model pattern
	parts := strings.SplitN(modelStr, "/", 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		prov := normalizeProviderAlias(parts[0])
		modelID := parts[1]
		if prov == "cloudflare-ai" && !strings.HasPrefix(modelID, "@cf/") && !strings.Contains(modelID, "/") {
			modelID = "@cf/" + modelID
		}
		return prov, modelID, true
	}
	alias, _ := repos.KVGet(db, "modelAliases", modelStr)
	if alias != nil {
		if s, ok := alias.(string); ok {
			parts = strings.SplitN(s, "/", 2)
			if len(parts) == 2 {
				return normalizeProviderAlias(parts[0]), parts[1], true
			}
		}
	}

	// Check customModels in KV
	rows, err := db.Query(`SELECT value FROM kv WHERE scope = 'customModels'`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var valStr string
			if err := rows.Scan(&valStr); err == nil {
				var cm struct {
					ID            string `json:"id"`
					Provider      string `json:"provider"`
					ProviderAlias string `json:"providerAlias"`
				}
				if json.Unmarshal([]byte(valStr), &cm) == nil {
					if cm.ID == modelStr {
						prov := cm.ProviderAlias
						if prov == "" {
							prov = cm.Provider
						}
						if prov != "" {
							return prov, modelStr, true
						}
					}
				}
			}
		}
	}

	// Well-known prefix heuristics for bare model names
	lower := strings.ToLower(modelStr)
	if strings.HasPrefix(lower, "gemini-") || strings.HasPrefix(lower, "gemma-") {
		return "gemini", modelStr, true
	}
	if strings.HasPrefix(lower, "claude-") {
		return "anthropic", modelStr, true
	}
	if strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3") || strings.HasPrefix(lower, "chatgpt") {
		return "openai", modelStr, true
	}
	if strings.HasPrefix(lower, "deepseek-") {
		return "deepseek", modelStr, true
	}
	if strings.HasPrefix(lower, "llama-") || strings.HasPrefix(lower, "mixtral") {
		return "groq", modelStr, true
	}
	if strings.HasPrefix(lower, "grok-") {
		return "xai", modelStr, true
	}
	if strings.HasPrefix(lower, "kimi-") {
		return "kimi", modelStr, true
	}
	if strings.HasPrefix(lower, "qwen") {
		return "together", modelStr, true
	}
	if strings.HasPrefix(lower, "bynara") || strings.HasPrefix(lower, "nry-") {
		return "bynara", modelStr, true
	}
	if strings.HasPrefix(lower, "geraikita") || strings.HasPrefix(lower, "gk/") {
		return "geraikita", modelStr, true
	}

	return "", modelStr, false
}
