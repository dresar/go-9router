package providers

import (
	"database/sql"
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
	return "", modelStr, false
}
