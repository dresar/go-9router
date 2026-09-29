package repos

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
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
	TotalRequests         int                      `json:"totalRequests"`
	TotalPromptTokens     int                      `json:"totalPromptTokens"`
	TotalCompletionTokens int                      `json:"totalCompletionTokens"`
	TotalCachedTokens     int                      `json:"totalCachedTokens"`
	TotalCost             float64                  `json:"totalCost"`
	TotalInputTokens      int                      `json:"totalInputTokens"`
	TotalOutputTokens     int                      `json:"totalOutputTokens"`
	ByProvider            map[string]ProviderStats `json:"byProvider"`
	ByModel               map[string]ModelStats    `json:"byModel"`
	ByAccount             map[string]AccountStats  `json:"byAccount"`
	ByApiKey              map[string]ApiKeyStats   `json:"byApiKey"`
	ByEndpoint            map[string]EndpointStats `json:"byEndpoint"`
	RecentRequests        []RecentRequestItem      `json:"recentRequests"`
	ActiveRequests        []any                    `json:"activeRequests"`
	Pending               map[string]any           `json:"pending"`
	Last10Minutes         []any                    `json:"last10Minutes"`
}

type ProviderStats struct {
	Requests         int     `json:"requests"`
	Cost             float64 `json:"cost"`
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	CachedTokens     int     `json:"cachedTokens"`
	InputTokens      int     `json:"inputTokens"`
	OutputTokens     int     `json:"outputTokens"`
	TotalTokens      int     `json:"totalTokens"`
}

type ModelStats struct {
	Requests         int     `json:"requests"`
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	CachedTokens     int     `json:"cachedTokens"`
	Cost             float64 `json:"cost"`
	RawModel         string  `json:"rawModel"`
	Provider         string  `json:"provider"`
	LastUsed         string  `json:"lastUsed"`
}

type AccountStats struct {
	Requests         int     `json:"requests"`
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	CachedTokens     int     `json:"cachedTokens"`
	Cost             float64 `json:"cost"`
	RawModel         string  `json:"rawModel"`
	Provider         string  `json:"provider"`
	ConnectionID     string  `json:"connectionId"`
	AccountName      string  `json:"accountName"`
	LastUsed         string  `json:"lastUsed"`
}

type ApiKeyStats struct {
	Requests         int     `json:"requests"`
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	CachedTokens     int     `json:"cachedTokens"`
	Cost             float64 `json:"cost"`
	RawModel         string  `json:"rawModel"`
	Provider         string  `json:"provider"`
	ApiKeyMasked     string  `json:"apiKeyMasked,omitempty"`
	KeyName          string  `json:"keyName"`
	ApiKeyKey        string  `json:"apiKeyKey"`
	LastUsed         string  `json:"lastUsed"`
}

type EndpointStats struct {
	Requests         int     `json:"requests"`
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	CachedTokens     int     `json:"cachedTokens"`
	Cost             float64 `json:"cost"`
	Endpoint         string  `json:"endpoint"`
	RawModel         string  `json:"rawModel"`
	Provider         string  `json:"provider"`
	LastUsed         string  `json:"lastUsed"`
}

type RecentRequestItem struct {
	Timestamp        string `json:"timestamp"`
	Model            string `json:"model"`
	Provider         string `json:"provider"`
	PromptTokens     int    `json:"promptTokens"`
	CompletionTokens int    `json:"completionTokens"`
	CachedTokens     int    `json:"cachedTokens"`
	Status           string `json:"status"`
}

type ChartBucket struct {
	Label    string  `json:"label"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
	Requests int     `json:"requests"`
}

func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 12 {
		return key[:1] + "***"
	}
	return key[:8] + "***" + key[len(key)-4:]
}

func GetUsageStats(db *sql.DB) (*UsageStats, error) {
	return GetUsageStatsWithPeriod(db, "today")
}

func GetUsageStatsWithPeriod(db *sql.DB, period string) (*UsageStats, error) {
	s := &UsageStats{
		ByProvider:     make(map[string]ProviderStats),
		ByModel:        make(map[string]ModelStats),
		ByAccount:      make(map[string]AccountStats),
		ByApiKey:       make(map[string]ApiKeyStats),
		ByEndpoint:     make(map[string]EndpointStats),
		RecentRequests: make([]RecentRequestItem, 0),
		ActiveRequests: make([]any, 0),
		Pending: map[string]any{
			"byModel":   map[string]int{},
			"byAccount": map[string]any{},
		},
		Last10Minutes: make([]any, 0),
	}

	// 1. Load connection names
	connMap := make(map[string]string)
	if cRows, err := db.Query(`SELECT id, name, email FROM providerConnections`); err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var id, name, email sql.NullString
			if cRows.Scan(&id, &name, &email) == nil && id.Valid {
				display := name.String
				if display == "" {
					display = email.String
				}
				if display == "" {
					display = id.String
				}
				connMap[id.String] = display
			}
		}
	}

	// 2. Load API keys
	apiKeyMap := make(map[string]string)
	if kRows, err := db.Query(`SELECT key, name FROM apiKeys`); err == nil {
		defer kRows.Close()
		for kRows.Next() {
			var key, name sql.NullString
			if kRows.Scan(&key, &name) == nil && key.Valid {
				display := name.String
				if display == "" {
					display = key.String
				}
				apiKeyMap[key.String] = display
			}
		}
	}

	// 3. Determine time cutoff based on period
	now := time.Now().UTC()
	var cutoff string
	switch strings.ToLower(period) {
	case "today":
		cutoff = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	case "24h":
		cutoff = now.Add(-24 * time.Hour).Format(time.RFC3339Nano)
	case "7d":
		cutoff = now.Add(-7 * 24 * time.Hour).Format(time.RFC3339Nano)
	case "30d":
		cutoff = now.Add(-30 * 24 * time.Hour).Format(time.RFC3339Nano)
	case "60d":
		cutoff = now.Add(-60 * 24 * time.Hour).Format(time.RFC3339Nano)
	case "all", "":
		cutoff = ""
	default:
		cutoff = ""
	}

	query := `SELECT id, timestamp, provider, model, connectionId, apiKey, endpoint, promptTokens, completionTokens, cost, status, tokens FROM usageHistory`
	args := []any{}
	if cutoff != "" {
		query += ` WHERE timestamp >= ?`
		args = append(args, cutoff)
	}
	query += ` ORDER BY id DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query usageHistory: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var ts, prov, mdl, connID, ak, ep, st, tk sql.NullString
		var pTokens, cTokens int
		var cost float64
		if err := rows.Scan(&id, &ts, &prov, &mdl, &connID, &ak, &ep, &pTokens, &cTokens, &cost, &st, &tk); err != nil {
			continue
		}

		timestamp := ts.String
		provider := prov.String
		if provider == "" {
			provider = "unknown"
		}
		model := mdl.String
		if model == "" {
			model = "unknown"
		}
		connectionID := connID.String
		apiKey := ak.String
		endpoint := ep.String
		if endpoint == "" {
			endpoint = "/v1/chat/completions"
		}
		status := st.String

		cachedTokens := 0
		if tk.Valid && tk.String != "" {
			var tokenObj map[string]any
			if json.Unmarshal([]byte(tk.String), &tokenObj) == nil {
				if ct, ok := tokenObj["cached_tokens"].(float64); ok {
					cachedTokens = int(ct)
				} else if ct, ok := tokenObj["cache_read_input_tokens"].(float64); ok {
					cachedTokens = int(ct)
				}
				if pTokens == 0 {
					if pt, ok := tokenObj["prompt_tokens"].(float64); ok {
						pTokens = int(pt)
					}
				}
				if cTokens == 0 {
					if cmpt, ok := tokenObj["completion_tokens"].(float64); ok {
						cTokens = int(cmpt)
					}
				}
			}
		}

		// Accumulate totals
		s.TotalRequests++
		s.TotalPromptTokens += pTokens
		s.TotalCompletionTokens += cTokens
		s.TotalCachedTokens += cachedTokens
		s.TotalCost += cost
		s.TotalInputTokens += pTokens
		s.TotalOutputTokens += cTokens

		// ByProvider
		ps := s.ByProvider[provider]
		ps.Requests++
		ps.PromptTokens += pTokens
		ps.CompletionTokens += cTokens
		ps.CachedTokens += cachedTokens
		ps.Cost += cost
		ps.InputTokens += pTokens
		ps.OutputTokens += cTokens
		ps.TotalTokens += pTokens + cTokens
		s.ByProvider[provider] = ps

		// ByModel
		modelKey := model
		if provider != "" && provider != "unknown" {
			modelKey = fmt.Sprintf("%s (%s)", model, provider)
		}
		ms, mExists := s.ByModel[modelKey]
		if !mExists {
			ms = ModelStats{
				RawModel: model,
				Provider: provider,
				LastUsed: timestamp,
			}
		}
		ms.Requests++
		ms.PromptTokens += pTokens
		ms.CompletionTokens += cTokens
		ms.CachedTokens += cachedTokens
		ms.Cost += cost
		if timestamp > ms.LastUsed {
			ms.LastUsed = timestamp
		}
		s.ByModel[modelKey] = ms

		// ByAccount
		if connectionID != "" {
			accountName := connMap[connectionID]
			if accountName == "" {
				if len(connectionID) > 8 {
					accountName = fmt.Sprintf("Account %s...", connectionID[:8])
				} else {
					accountName = fmt.Sprintf("Account %s", connectionID)
				}
			}
			accountKey := fmt.Sprintf("%s (%s - %s)", model, provider, accountName)
			as, aExists := s.ByAccount[accountKey]
			if !aExists {
				as = AccountStats{
					RawModel:     model,
					Provider:     provider,
					ConnectionID: connectionID,
					AccountName:  accountName,
					LastUsed:     timestamp,
				}
			}
			as.Requests++
			as.PromptTokens += pTokens
			as.CompletionTokens += cTokens
			as.CachedTokens += cachedTokens
			as.Cost += cost
			if timestamp > as.LastUsed {
				as.LastUsed = timestamp
			}
			s.ByAccount[accountKey] = as
		}

		// ByApiKey
		var akKey string
		if apiKey != "" {
			akKey = fmt.Sprintf("%s|%s|%s", apiKey, model, provider)
			keyName := apiKeyMap[apiKey]
			if keyName == "" {
				if len(apiKey) > 8 {
					keyName = apiKey[:8] + "..."
				} else {
					keyName = apiKey
				}
			}
			masked := maskAPIKey(apiKey)
			aks, akExists := s.ByApiKey[akKey]
			if !akExists {
				aks = ApiKeyStats{
					RawModel:     model,
					Provider:     provider,
					ApiKeyMasked: masked,
					KeyName:      keyName,
					ApiKeyKey:    masked,
					LastUsed:     timestamp,
				}
			}
			aks.Requests++
			aks.PromptTokens += pTokens
			aks.CompletionTokens += cTokens
			aks.CachedTokens += cachedTokens
			aks.Cost += cost
			if timestamp > aks.LastUsed {
				aks.LastUsed = timestamp
			}
			s.ByApiKey[akKey] = aks
		} else {
			akKey = "local-no-key"
			aks, akExists := s.ByApiKey[akKey]
			if !akExists {
				aks = ApiKeyStats{
					RawModel:  model,
					Provider:  provider,
					KeyName:   "Local (No API Key)",
					ApiKeyKey: "local-no-key",
					LastUsed:  timestamp,
				}
			}
			aks.Requests++
			aks.PromptTokens += pTokens
			aks.CompletionTokens += cTokens
			aks.CachedTokens += cachedTokens
			aks.Cost += cost
			if timestamp > aks.LastUsed {
				aks.LastUsed = timestamp
			}
			s.ByApiKey[akKey] = aks
		}

		// ByEndpoint
		epKey := fmt.Sprintf("%s|%s|%s", endpoint, model, provider)
		eps, epExists := s.ByEndpoint[epKey]
		if !epExists {
			eps = EndpointStats{
				Endpoint: endpoint,
				RawModel: model,
				Provider: provider,
				LastUsed: timestamp,
			}
		}
		eps.Requests++
		eps.PromptTokens += pTokens
		eps.CompletionTokens += cTokens
		eps.CachedTokens += cachedTokens
		eps.Cost += cost
		if timestamp > eps.LastUsed {
			eps.LastUsed = timestamp
		}
		s.ByEndpoint[epKey] = eps

		// RecentRequests (most recent 20)
		if len(s.RecentRequests) < 20 && (pTokens > 0 || cTokens > 0) {
			statusVal := status
			if statusVal == "success" || statusVal == "" {
				statusVal = "ok"
			}
			s.RecentRequests = append(s.RecentRequests, RecentRequestItem{
				Timestamp:        timestamp,
				Model:            model,
				Provider:         provider,
				PromptTokens:     pTokens,
				CompletionTokens: cTokens,
				CachedTokens:     cachedTokens,
				Status:           statusVal,
			})
		}
	}

	return s, nil
}

func GetRecentLogs(db *sql.DB, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	conns, _ := ListConnections(db, ConnectionFilter{})
	connMap := make(map[string]string)
	for _, c := range conns {
		name := c.Name
		if name == "" {
			name = c.Email
		}
		connMap[c.ID] = name
	}

	rows, err := db.Query(`SELECT timestamp, provider, model, connectionId, promptTokens, completionTokens, status, tokens FROM usageHistory ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var ts, prov, mdl, connID, st, tk sql.NullString
		var pTokens, cTokens int
		if err := rows.Scan(&ts, &prov, &mdl, &connID, &pTokens, &cTokens, &st, &tk); err != nil {
			continue
		}

		formattedDate := ts.String
		if t, err := time.Parse(time.RFC3339Nano, ts.String); err == nil {
			formattedDate = t.Format("02-01-2006 15:04:05")
		} else if t, err := time.Parse(time.RFC3339, ts.String); err == nil {
			formattedDate = t.Format("02-01-2006 15:04:05")
		}

		p := strings.ToUpper(prov.String)
		if p == "" {
			p = "-"
		}
		m := mdl.String
		if m == "" {
			m = "-"
		}
		acc := connMap[connID.String]
		if acc == "" {
			if connID.String != "" {
				if len(connID.String) > 8 {
					acc = connID.String[:8]
				} else {
					acc = connID.String
				}
			} else {
				acc = "-"
			}
		}

		statusStr := st.String
		if statusStr == "success" || statusStr == "ok" || statusStr == "" {
			statusStr = "200 OK"
		} else if statusStr == "pending" {
			statusStr = "PENDING"
		} else if strings.HasPrefix(statusStr, "4") || strings.HasPrefix(statusStr, "5") {
			statusStr = statusStr + " FAILED"
		}

		line := fmt.Sprintf("%s | %s | %s | %s | %d | %d | %s",
			formattedDate, m, p, acc, pTokens, cTokens, statusStr)
		out = append(out, line)
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func GetChartData(db *sql.DB, period string) ([]ChartBucket, error) {
	now := time.Now().UTC()

	switch strings.ToLower(period) {
	case "today":
		bucketCount := 24
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		buckets := make([]ChartBucket, bucketCount)
		for i := 0; i < bucketCount; i++ {
			t := startOfDay.Add(time.Duration(i) * time.Hour)
			buckets[i] = ChartBucket{
				Label:    t.Format("15:04"),
				Tokens:   0,
				Cost:     0,
				Requests: 0,
			}
		}

		rows, err := db.Query(`SELECT timestamp, promptTokens, completionTokens, cost FROM usageHistory WHERE timestamp >= ?`, startOfDay.Format(time.RFC3339Nano))
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var ts string
				var pTokens, cTokens int
				var cost float64
				if rows.Scan(&ts, &pTokens, &cTokens, &cost) == nil {
					t, err := time.Parse(time.RFC3339Nano, ts)
					if err != nil {
						t, _ = time.Parse(time.RFC3339, ts)
					}
					idx := int(t.Sub(startOfDay).Hours())
					if idx >= 0 && idx < bucketCount {
						buckets[idx].Tokens += int64(pTokens + cTokens)
						buckets[idx].Cost += cost
						buckets[idx].Requests++
					}
				}
			}
		}
		return buckets, nil

	case "24h":
		bucketCount := 24
		startTime := now.Add(-23 * time.Hour).Truncate(time.Hour)
		buckets := make([]ChartBucket, bucketCount)
		for i := 0; i < bucketCount; i++ {
			t := startTime.Add(time.Duration(i) * time.Hour)
			buckets[i] = ChartBucket{
				Label:    t.Format("15:04"),
				Tokens:   0,
				Cost:     0,
				Requests: 0,
			}
		}

		rows, err := db.Query(`SELECT timestamp, promptTokens, completionTokens, cost FROM usageHistory WHERE timestamp >= ?`, startTime.Format(time.RFC3339Nano))
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var ts string
				var pTokens, cTokens int
				var cost float64
				if rows.Scan(&ts, &pTokens, &cTokens, &cost) == nil {
					t, err := time.Parse(time.RFC3339Nano, ts)
					if err != nil {
						t, _ = time.Parse(time.RFC3339, ts)
					}
					idx := int(t.Sub(startTime).Hours())
					if idx >= 0 && idx < bucketCount {
						buckets[idx].Tokens += int64(pTokens + cTokens)
						buckets[idx].Cost += cost
						buckets[idx].Requests++
					}
				}
			}
		}
		return buckets, nil

	case "7d", "30d", "60d", "all":
		days := 7
		if strings.ToLower(period) == "30d" {
			days = 30
		} else if strings.ToLower(period) == "60d" {
			days = 60
		} else if strings.ToLower(period) == "all" {
			days = 30
			var earliest string
			_ = db.QueryRow(`SELECT MIN(timestamp) FROM usageHistory`).Scan(&earliest)
			if earliest != "" {
				if et, err := time.Parse(time.RFC3339Nano, earliest); err == nil {
					diff := int(now.Sub(et).Hours()/24) + 1
					if diff > 1 && diff <= 365 {
						days = diff
					}
				}
			}
		}

		startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(days - 1))
		buckets := make([]ChartBucket, days)
		for i := 0; i < days; i++ {
			d := startTime.AddDate(0, 0, i)
			buckets[i] = ChartBucket{
				Label:    d.Format("Jan 02"),
				Tokens:   0,
				Cost:     0,
				Requests: 0,
			}
		}

		rows, err := db.Query(`SELECT timestamp, promptTokens, completionTokens, cost FROM usageHistory WHERE timestamp >= ?`, startTime.Format(time.RFC3339Nano))
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var ts string
				var pTokens, cTokens int
				var cost float64
				if rows.Scan(&ts, &pTokens, &cTokens, &cost) == nil {
					t, err := time.Parse(time.RFC3339Nano, ts)
					if err != nil {
						t, _ = time.Parse(time.RFC3339, ts)
					}
					idx := int(t.Sub(startTime).Hours() / 24)
					if idx >= 0 && idx < days {
						buckets[idx].Tokens += int64(pTokens + cTokens)
						buckets[idx].Cost += cost
						buckets[idx].Requests++
					}
				}
			}
		}
		return buckets, nil

	default:
		return GetChartData(db, "24h")
	}
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
