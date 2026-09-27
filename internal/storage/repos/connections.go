package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Connection struct {
	ID        string         `json:"id"`
	Provider  string         `json:"provider"`
	AuthType  string         `json:"authType"`
	Name      string         `json:"name,omitempty"`
	Email     string         `json:"email,omitempty"`
	Priority  int            `json:"priority"`
	IsActive  bool           `json:"isActive"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
	Data      map[string]any `json:"-"`

	APIKey              string `json:"apiKey,omitempty"`
	AccessToken         string `json:"accessToken,omitempty"`
	RefreshToken        string `json:"refreshToken,omitempty"`
	IDToken             string `json:"idToken,omitempty"`
	ExpiresAt           string `json:"expiresAt,omitempty"`
	ExpiresIn           int    `json:"expiresIn,omitempty"`
	LastRefreshAt       string `json:"lastRefreshAt,omitempty"`
	ProjectID           string `json:"projectId,omitempty"`
	TestStatus          string `json:"testStatus,omitempty"`
	LastError           string `json:"lastError,omitempty"`
	ErrorCode           int    `json:"errorCode,omitempty"`
	LastErrorAt         string `json:"lastErrorAt,omitempty"`
	BackoffLevel        int    `json:"backoffLevel,omitempty"`
	LastUsedAt          string `json:"lastUsedAt,omitempty"`
	ConsecutiveUseCount int    `json:"consecutiveUseCount,omitempty"`
	DisplayName         string `json:"displayName,omitempty"`

	ProviderSpecificData map[string]any `json:"providerSpecificData,omitempty"`
}

type ConnectionFilter struct {
	Provider *string
	IsActive *bool
}

func ListConnections(db *sql.DB, f ConnectionFilter) ([]Connection, error) {
	q := `SELECT id, provider, authType, name, email, priority, isActive, data, createdAt, updatedAt FROM providerConnections WHERE 1=1`
	args := []any{}
	if f.Provider != nil {
		q += ` AND provider = ?`
		args = append(args, *f.Provider)
	}
	if f.IsActive != nil {
		v := 0
		if *f.IsActive {
			v = 1
		}
		q += ` AND isActive = ?`
		args = append(args, v)
	}
	q += ` ORDER BY priority ASC, createdAt ASC`
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}
	defer rows.Close()
	var out []Connection
	for rows.Next() {
		c, err := scanConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func GetConnection(db *sql.DB, id string) (*Connection, error) {
	row := db.QueryRow(`SELECT id, provider, authType, name, email, priority, isActive, data, createdAt, updatedAt FROM providerConnections WHERE id = ?`, id)
	c, err := scanConnection(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CreateConnection(db *sql.DB, c Connection) (*Connection, error) {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	c.CreatedAt = now
	c.UpdatedAt = now
	data := connectionToData(c)
	b, _ := json.Marshal(data)
	isActive := 1
	if !c.IsActive {
		isActive = 0
	}
	_, err := db.Exec(
		`INSERT INTO providerConnections(id, provider, authType, name, email, priority, isActive, data, createdAt, updatedAt) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		c.ID, c.Provider, c.AuthType, nvl(c.Name), nvl(c.Email), c.Priority, isActive, string(b), c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}
	return &c, nil
}

func UpdateConnection(db *sql.DB, id string, updates map[string]any) (*Connection, error) {
	c, err := GetConnection(db, id)
	if err != nil || c == nil {
		return nil, err
	}
	data := connectionToData(*c)
	for k, v := range updates {
		if v == nil {
			delete(data, k)
		} else {
			data[k] = v
		}
	}
	updated := dataToConnection(id, c.Provider, c.AuthType, c.Name, c.Email, c.Priority, c.IsActive, data, c.CreatedAt)
	updated.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	b, _ := json.Marshal(connectionToData(updated))
	isActive := 1
	if !updated.IsActive {
		isActive = 0
	}
	_, err = db.Exec(
		`UPDATE providerConnections SET authType=?, name=?, email=?, priority=?, isActive=?, data=?, updatedAt=? WHERE id=?`,
		updated.AuthType, nvl(updated.Name), nvl(updated.Email), updated.Priority, isActive, string(b), updated.UpdatedAt, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update connection: %w", err)
	}
	return &updated, nil
}

func DeleteConnection(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM providerConnections WHERE id = ?`, id)
	return err
}

func scanConnection(row interface{ Scan(...any) error }) (Connection, error) {
	var id, provider, authType, createdAt, updatedAt string
	var name, email sql.NullString
	var priority, isActive int
	var dataStr string
	if err := row.Scan(&id, &provider, &authType, &name, &email, &priority, &isActive, &dataStr, &createdAt, &updatedAt); err != nil {
		return Connection{}, err
	}
	var data map[string]any
	_ = json.Unmarshal([]byte(dataStr), &data)
	c := dataToConnection(id, provider, authType, name.String, email.String, priority, isActive == 1, data, createdAt)
	c.UpdatedAt = updatedAt
	return c, nil
}

func connectionToData(c Connection) map[string]any {
	m := map[string]any{}
	if c.APIKey != "" {
		m["apiKey"] = c.APIKey
	}
	if c.AccessToken != "" {
		m["accessToken"] = c.AccessToken
	}
	if c.RefreshToken != "" {
		m["refreshToken"] = c.RefreshToken
	}
	if c.IDToken != "" {
		m["idToken"] = c.IDToken
	}
	if c.ExpiresAt != "" {
		m["expiresAt"] = c.ExpiresAt
	}
	if c.ExpiresIn != 0 {
		m["expiresIn"] = c.ExpiresIn
	}
	if c.LastRefreshAt != "" {
		m["lastRefreshAt"] = c.LastRefreshAt
	}
	if c.ProjectID != "" {
		m["projectId"] = c.ProjectID
	}
	if c.TestStatus != "" {
		m["testStatus"] = c.TestStatus
	}
	if c.LastError != "" {
		m["lastError"] = c.LastError
	}
	if c.ErrorCode != 0 {
		m["errorCode"] = c.ErrorCode
	}
	if c.LastErrorAt != "" {
		m["lastErrorAt"] = c.LastErrorAt
	}
	if c.BackoffLevel != 0 {
		m["backoffLevel"] = c.BackoffLevel
	}
	if c.LastUsedAt != "" {
		m["lastUsedAt"] = c.LastUsedAt
	}
	if c.ConsecutiveUseCount != 0 {
		m["consecutiveUseCount"] = c.ConsecutiveUseCount
	}
	if c.DisplayName != "" {
		m["displayName"] = c.DisplayName
	}
	if c.ProviderSpecificData != nil {
		m["providerSpecificData"] = c.ProviderSpecificData
	}
	for k, v := range c.Data {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}
	return m
}

func dataToConnection(id, provider, authType, name, email string, priority int, isActive bool, data map[string]any, createdAt string) Connection {
	c := Connection{
		ID:        id,
		Provider:  provider,
		AuthType:  authType,
		Name:      name,
		Email:     email,
		Priority:  priority,
		IsActive:  isActive,
		CreatedAt: createdAt,
		Data:      data,
	}
	if data == nil {
		return c
	}
	c.APIKey, _ = data["apiKey"].(string)
	c.AccessToken, _ = data["accessToken"].(string)
	c.RefreshToken, _ = data["refreshToken"].(string)
	c.IDToken, _ = data["idToken"].(string)
	c.ExpiresAt, _ = data["expiresAt"].(string)
	if v, ok := data["expiresIn"].(float64); ok {
		c.ExpiresIn = int(v)
	}
	c.LastRefreshAt, _ = data["lastRefreshAt"].(string)
	c.ProjectID, _ = data["projectId"].(string)
	c.TestStatus, _ = data["testStatus"].(string)
	c.LastError, _ = data["lastError"].(string)
	if v, ok := data["errorCode"].(float64); ok {
		c.ErrorCode = int(v)
	}
	c.LastErrorAt, _ = data["lastErrorAt"].(string)
	if v, ok := data["backoffLevel"].(float64); ok {
		c.BackoffLevel = int(v)
	}
	c.LastUsedAt, _ = data["lastUsedAt"].(string)
	if v, ok := data["consecutiveUseCount"].(float64); ok {
		c.ConsecutiveUseCount = int(v)
	}
	c.DisplayName, _ = data["displayName"].(string)
	if psd, ok := data["providerSpecificData"].(map[string]any); ok {
		c.ProviderSpecificData = psd
	}
	return c
}

func nvl(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
