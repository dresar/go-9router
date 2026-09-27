package providers

import (
	"database/sql"
	"encoding/json"
	"strings"
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
	Credentials  *Credentials
	AllLocked    bool
	RetryAfter   string
}

func SelectCredentials(db *sql.DB, provider string, exclude map[string]bool) (*SelectionResult, error) {
	active := true
	conns, err := repos.ListConnections(db, repos.ConnectionFilter{
		Provider: &provider,
		IsActive: &active,
	})
	if err != nil {
		return nil, err
	}

	if len(conns) == 0 {
		return nil, nil
	}

	available := make([]repos.Connection, 0, len(conns))
	for _, c := range conns {
		if exclude[c.ID] {
			continue
		}
		if isModelLocked(c) {
			continue
		}
		available = append(available, c)
	}

	if len(available) == 0 {
		return &SelectionResult{AllLocked: true}, nil
	}

	conn := available[0]
	return &SelectionResult{
		Credentials: connectionToCredentials(conn),
	}, nil
}

func isModelLocked(c repos.Connection) bool {
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
	return time.Since(t) < 60*time.Second
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
	return err
}

func ClearError(db *sql.DB, connectionID string) error {
	_, err := repos.UpdateConnection(db, connectionID, map[string]any{
		"testStatus":  "active",
		"lastError":   nil,
		"errorCode":   nil,
		"lastErrorAt": nil,
	})
	return err
}

func ResolveModelProvider(modelStr string, db *sql.DB) (provider, model string, ok bool) {
	if modelStr == "" {
		return "", "", false
	}
	parts := strings.SplitN(modelStr, "/", 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], true
	}
	alias, _ := repos.KVGet(db, "modelAliases", modelStr)
	if alias != nil {
		if s, ok := alias.(string); ok {
			parts = strings.SplitN(s, "/", 2)
			if len(parts) == 2 {
				return parts[0], parts[1], true
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

	return "", modelStr, false
}

