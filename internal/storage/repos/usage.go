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
	ID               string         `json:"id"`
	Timestamp        string         `json:"timestamp"`
	Provider         string         `json:"provider,omitempty"`
	Model            string         `json:"model,omitempty"`
	ConnectionID     string         `json:"connectionId,omitempty"`
	Status           string         `json:"status,omitempty"`
	Latency          map[string]any `json:"latency,omitempty"`
	Tokens           map[string]any `json:"tokens,omitempty"`
	Request          any            `json:"request,omitempty"`
	ProviderRequest  any            `json:"providerRequest,omitempty"`
	ProviderResponse any            `json:"providerResponse,omitempty"`
	Response         any            `json:"response,omitempty"`
	Data             map[string]any `json:"data,omitempty"`
}

type RequestDetailFilter struct {
	Provider     string
	Model        string
	ConnectionID string
	Status       string
	StartDate    string
	EndDate      string
	Limit        int
	Offset       int
}

func SaveRequestDetail(db *sql.DB, d RequestDetail) error {
	if d.Timestamp == "" {
		d.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	m := make(map[string]any)
	if d.Data != nil {
		for k, v := range d.Data {
			m[k] = v
		}
	}
	m["id"] = d.ID
	m["timestamp"] = d.Timestamp
	if d.Provider != "" {
		m["provider"] = d.Provider
	}
	if d.Model != "" {
		m["model"] = d.Model
	}
	if d.ConnectionID != "" {
		m["connectionId"] = d.ConnectionID
	}
	if d.Status != "" {
		m["status"] = d.Status
	}
	if d.Latency != nil {
		m["latency"] = d.Latency
	}
	if d.Tokens != nil {
		m["tokens"] = d.Tokens
	}
	if d.Request != nil {
		m["request"] = d.Request
	}
	if d.ProviderRequest != nil {
		m["providerRequest"] = d.ProviderRequest
	}
	if d.ProviderResponse != nil {
		m["providerResponse"] = d.ProviderResponse
	}
	if d.Response != nil {
		m["response"] = d.Response
	}
	b, _ := json.Marshal(m)
	_, err := db.Exec(
		`INSERT OR REPLACE INTO requestDetails(id, timestamp, provider, model, connectionId, status, data) VALUES(?,?,?,?,?,?,?)`,
		d.ID, d.Timestamp, nvl(d.Provider), nvl(d.Model), nvl(d.ConnectionID), nvl(d.Status), string(b),
	)
	return err
}

func ListRequestDetailsFiltered(db *sql.DB, f RequestDetailFilter) ([]map[string]any, int, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT data FROM requestDetails WHERE 1=1`
	cq := `SELECT COUNT(*) FROM requestDetails WHERE 1=1`
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
	if f.Status != "" {
		q += ` AND status = ?`
		cq += ` AND status = ?`
		args = append(args, f.Status)
	}
	if f.StartDate != "" {
		q += ` AND timestamp >= ?`
		cq += ` AND timestamp >= ?`
		args = append(args, f.StartDate)
	}
	if f.EndDate != "" {
		q += ` AND timestamp <= ?`
		cq += ` AND timestamp <= ?`
		args = append(args, f.EndDate)
	}

	var total int
	_ = db.QueryRow(cq, args...).Scan(&total)

	q += ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, f.Offset)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var dataStr string
		if err := rows.Scan(&dataStr); err != nil {
			continue
		}
		var item map[string]any
		if err := json.Unmarshal([]byte(dataStr), &item); err == nil {
			out = append(out, item)
		}
	}
	if out == nil {
		out = []map[string]any{}
	}
	return out, total, rows.Err()
}

func ListRequestDetails(db *sql.DB, limit, offset int) ([]map[string]any, error) {
	details, _, err := ListRequestDetailsFiltered(db, RequestDetailFilter{Limit: limit, Offset: offset})
	return details, err
}
