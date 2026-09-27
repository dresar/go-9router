package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type UsageRecord struct {
	ID               int64   `json:"id"`
	Timestamp        string  `json:"timestamp"`
	Provider         string  `json:"provider,omitempty"`
	Model            string  `json:"model,omitempty"`
	ConnectionID     string  `json:"connectionId,omitempty"`
	APIKey           string  `json:"apiKey,omitempty"`
	Endpoint         string  `json:"endpoint,omitempty"`
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	Cost             float64 `json:"cost"`
	Status           string  `json:"status,omitempty"`
	Tokens           string  `json:"tokens,omitempty"`
	Meta             string  `json:"meta,omitempty"`
}

type UsageFilter struct {
	Provider     string
	Model        string
	ConnectionID string
	Limit        int
	Offset       int
}

func SaveUsage(db *sql.DB, r UsageRecord) error {
	if r.Timestamp == "" {
		r.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	_, err := db.Exec(
		`INSERT INTO usageHistory(timestamp, provider, model, connectionId, apiKey, endpoint, promptTokens, completionTokens, cost, status, tokens, meta)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.Timestamp, nvl(r.Provider), nvl(r.Model), nvl(r.ConnectionID), nvl(r.APIKey), nvl(r.Endpoint),
		r.PromptTokens, r.CompletionTokens, r.Cost, nvl(r.Status), nvl(r.Tokens), nvl(r.Meta),
	)
	return err
}

func ListUsage(db *sql.DB, f UsageFilter) ([]UsageRecord, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT id, timestamp, provider, model, connectionId, apiKey, endpoint, promptTokens, completionTokens, cost, status, tokens, meta FROM usageHistory WHERE 1=1`
	cq := `SELECT COUNT(*) FROM usageHistory WHERE 1=1`
	args := []any{}
	if f.Provider != "" {
		q += ` AND provider = ?`
		cq += ` AND provider = ?`
		args = append(args, f.Provider)
	}
	if f.Model != "" {
		q += ` AND model = ?`
		cq += ` AND model = ?`
		args = append(args, f.Model)
	}
	if f.ConnectionID != "" {
		q += ` AND connectionId = ?`
		cq += ` AND connectionId = ?`
		args = append(args, f.ConnectionID)
	}
	var total int
	_ = db.QueryRow(cq, args...).Scan(&total)
	q += ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, f.Offset)
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list usage: %w", err)
	}
	defer rows.Close()
	var out []UsageRecord
	for rows.Next() {
		var r UsageRecord
		var provider, model, connID, apiKey, endpoint, status, tokens, meta sql.NullString
		if err := rows.Scan(&r.ID, &r.Timestamp, &provider, &model, &connID, &apiKey, &endpoint, &r.PromptTokens, &r.CompletionTokens, &r.Cost, &status, &tokens, &meta); err != nil {
			return nil, 0, err
		}
		r.Provider = provider.String
		r.Model = model.String
		r.ConnectionID = connID.String
		r.APIKey = apiKey.String
		r.Endpoint = endpoint.String
		r.Status = status.String
		r.Tokens = tokens.String
		r.Meta = meta.String
		out = append(out, r)
	}
	return out, total, rows.Err()
}

type UsageStats struct {
	TotalRequests    int     `json:"totalRequests"`
	TotalCost        float64 `json:"totalCost"`
	TotalInputTokens int     `json:"totalInputTokens"`
	TotalOutputTokens int    `json:"totalOutputTokens"`
	ByProvider       map[string]ProviderStats `json:"byProvider"`
}

type ProviderStats struct {
	Requests         int     `json:"requests"`
	Cost             float64 `json:"cost"`
	InputTokens      int     `json:"inputTokens"`
	OutputTokens     int     `json:"outputTokens"`
}

func GetUsageStats(db *sql.DB) (*UsageStats, error) {
	s := &UsageStats{ByProvider: map[string]ProviderStats{}}
	row := db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(cost),0), COALESCE(SUM(promptTokens),0), COALESCE(SUM(completionTokens),0) FROM usageHistory`)
	if err := row.Scan(&s.TotalRequests, &s.TotalCost, &s.TotalInputTokens, &s.TotalOutputTokens); err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT provider, COUNT(*), COALESCE(SUM(cost),0), COALESCE(SUM(promptTokens),0), COALESCE(SUM(completionTokens),0) FROM usageHistory GROUP BY provider`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var provider sql.NullString
		var ps ProviderStats
		if err := rows.Scan(&provider, &ps.Requests, &ps.Cost, &ps.InputTokens, &ps.OutputTokens); err != nil {
			return nil, err
		}
		s.ByProvider[provider.String] = ps
	}
	return s, rows.Err()
}

type RequestDetail struct {
	ID           string         `json:"id"`
	Timestamp    string         `json:"timestamp"`
	Provider     string         `json:"provider,omitempty"`
	Model        string         `json:"model,omitempty"`
	ConnectionID string         `json:"connectionId,omitempty"`
	Status       string         `json:"status,omitempty"`
	Data         map[string]any `json:"data,omitempty"`
}

func SaveRequestDetail(db *sql.DB, d RequestDetail) error {
	if d.Timestamp == "" {
		d.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	b, _ := json.Marshal(d.Data)
	_, err := db.Exec(
		`INSERT OR REPLACE INTO requestDetails(id, timestamp, provider, model, connectionId, status, data) VALUES(?,?,?,?,?,?,?)`,
		d.ID, d.Timestamp, nvl(d.Provider), nvl(d.Model), nvl(d.ConnectionID), nvl(d.Status), string(b),
	)
	return err
}

func ListRequestDetails(db *sql.DB, limit, offset int) ([]RequestDetail, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.Query(
		`SELECT id, timestamp, provider, model, connectionId, status, data FROM requestDetails ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RequestDetail
	for rows.Next() {
		var d RequestDetail
		var provider, model, connID, status sql.NullString
		var dataStr string
		if err := rows.Scan(&d.ID, &d.Timestamp, &provider, &model, &connID, &status, &dataStr); err != nil {
			return nil, err
		}
		d.Provider = provider.String
		d.Model = model.String
		d.ConnectionID = connID.String
		d.Status = status.String
		_ = json.Unmarshal([]byte(dataStr), &d.Data)
		out = append(out, d)
	}
	return out, rows.Err()
}
