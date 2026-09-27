package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dresar/go-9router/internal/providers"
	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleUsageStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	stats, err := repos.GetUsageStats(h.DB)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to get usage stats")
		return
	}
	h.JSON(w, http.StatusOK, stats)
}

func (h *Handler) HandleUsageHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	q := r.URL.Query()
	filter := repos.UsageFilter{
		Provider:     q.Get("provider"),
		Model:        q.Get("model"),
		ConnectionID: q.Get("connectionId"),
	}
	records, total, err := repos.ListUsage(h.DB, filter)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to get usage history")
		return
	}
	if records == nil {
		records = []repos.UsageRecord{}
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"history": records,
		"total":   total,
	})
}

func (h *Handler) HandleUsageStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.JSONError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}
	w.Write([]byte("data: {\"type\":\"connected\"}\n\n"))
	flusher.Flush()
	<-r.Context().Done()
}

func (h *Handler) HandleRequestDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	filter := repos.RequestDetailFilter{
		Provider:  q.Get("provider"),
		Model:     q.Get("model"),
		StartDate: q.Get("startDate"),
		EndDate:   q.Get("endDate"),
		Limit:     pageSize,
		Offset:    offset,
	}
	details, total, err := repos.ListRequestDetailsFiltered(h.DB, filter)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to get request details")
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	h.JSON(w, http.StatusOK, map[string]any{
		"details": details,
		"pagination": map[string]any{
			"page":       page,
			"pageSize":   pageSize,
			"totalItems": total,
			"totalPages": totalPages,
			"hasNext":    page < totalPages,
			"hasPrev":    page > 1,
		},
	})
}

func (h *Handler) HandleRequestLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	details, err := repos.ListRequestDetails(h.DB, 100, 0)
	if err != nil || details == nil {
		h.JSON(w, http.StatusOK, []any{})
		return
	}
	h.JSON(w, http.StatusOK, details)
}

func (h *Handler) HandleUsageChart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	now := time.Now()
	type Bucket struct {
		Label    string  `json:"label"`
		Tokens   int64   `json:"tokens"`
		Cost     float64 `json:"cost"`
		Requests int     `json:"requests"`
	}
	buckets := make([]Bucket, 24)
	for i := 0; i < 24; i++ {
		t := now.Add(-time.Duration(23-i) * time.Hour)
		buckets[i] = Bucket{
			Label:    t.Format("15:04"),
			Tokens:   0,
			Cost:     0,
			Requests: 0,
		}
	}
	h.JSON(w, http.StatusOK, buckets)
}

func (h *Handler) HandleUsageProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	conns, _ := repos.ListConnections(h.DB, repos.ConnectionFilter{})
	seen := make(map[string]bool)
	var list []map[string]string
	for _, c := range conns {
		if c.Provider != "" && !seen[c.Provider] {
			seen[c.Provider] = true
			list = append(list, map[string]string{"id": c.Provider, "name": c.Provider})
		}
	}
	if len(list) == 0 {
		list = []map[string]string{
			{"id": "claude", "name": "Claude"},
			{"id": "openai", "name": "OpenAI"},
			{"id": "gemini", "name": "Gemini"},
			{"id": "deepseek", "name": "DeepSeek"},
		}
	}
	h.JSON(w, http.StatusOK, map[string]any{"providers": list})
}

func (h *Handler) HandleUsageLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{"logs": []any{}})
}

func (h *Handler) HandleUsageConnectionSub(w http.ResponseWriter, r *http.Request) {
	sub := pathSegment(r, "/api/usage/")
	if strings.Contains(sub, "reset") {
		h.JSON(w, http.StatusOK, map[string]any{
			"ok":            true,
			"reset":         true,
			"code":          "success",
			"windows_reset": time.Now().Add(5 * time.Hour).Unix(),
		})
		return
	}

	// 1. If Next.js / Node.js backend is active on port 20127, proxy so provider-specific quota engines run
	if isPortOpen("127.0.0.1", 20127) {
		target, err := url.Parse("http://127.0.0.1:20127")
		if err == nil {
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.ServeHTTP(w, r)
			return
		}
	}

	conn, err := repos.GetConnection(h.DB, sub)
	if err != nil || conn == nil {
		h.HandleUsageStats(w, r)
		return
	}

	quotas := map[string]any{}
	plan := "Pro"
	var message string

	if strings.ToLower(conn.Provider) == "antigravity" && conn.AccessToken != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		loadReqBody := []byte(`{"project":"` + conn.ProjectID + `"}`)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://daily-cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels", bytes.NewReader(loadReqBody))
		if req != nil {
			req.Header.Set("Authorization", "Bearer "+conn.AccessToken)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "antigravity/ide/2.11.0 windows/amd64")
			req.Header.Set("X-Client-Name", "antigravity")
			req.Header.Set("X-Client-Version", "2.11.0")

			res, uErr := providers.DoUpstreamWithProxy(ctx, req, h.DB, &providers.Credentials{ProxyPoolID: conn.ProxyPoolID})
			if uErr == nil && res.Status == 200 && res.Response != nil {
				defer res.Response.Body.Close()
				var data struct {
					Models map[string]any `json:"models"`
				}
				if json.NewDecoder(res.Response.Body).Decode(&data) == nil && len(data.Models) > 0 {
					for k, v := range data.Models {
						if m, ok := v.(map[string]any); ok {
							quotas[k] = m
						}
					}
				}
			}
		}
	}

	if len(quotas) == 0 {
		message = "Quota API tidak disediakan oleh provider ini"
	}

	resp := map[string]any{
		"plan":   plan,
		"quotas": quotas,
	}
	if message != "" {
		resp["message"] = message
	}

	h.JSON(w, http.StatusOK, resp)
}
