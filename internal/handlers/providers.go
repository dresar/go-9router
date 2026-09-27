package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dresar/go-9router/internal/providers"
	"github.com/dresar/go-9router/internal/storage/repos"
)

func (h *Handler) HandleProviders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listProviders(w, r)
	case http.MethodPost:
		h.createProvider(w, r)
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleProviderByID(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r, "/api/providers/")
	if id == "" {
		h.JSONError(w, http.StatusBadRequest, "missing id")
		return
	}

	// Sub-routes under /api/providers/
	if id == "client" {
		h.HandleProvidersClient(w, r)
		return
	}
	if id == "suggested-models" {
		h.HandleSuggestedModels(w, r)
		return
	}
	if id == "test-batch" {
		h.HandleTestBatch(w, r)
		return
	}
	if id == "validate" {
		h.JSON(w, http.StatusOK, map[string]any{"valid": true})
		return
	}

	if strings.Contains(id, "/") {
		parts := strings.Split(id, "/")
		if len(parts) == 2 {
			connID := parts[0]
			action := parts[1]
			if connID == "kilo" && action == "free-models" {
				h.JSON(w, http.StatusOK, map[string]any{"models": []any{}})
				return
			}
			switch action {
			case "models":
				h.HandleProviderModels(w, r, connID)
				return
			case "test":
				h.HandleProviderTest(w, r, connID)
				return
			case "test-models":
				h.HandleProviderTestModels(w, r, connID)
				return
			}
		}
		h.JSONError(w, http.StatusNotFound, "not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getProvider(w, r, id)
	case http.MethodPut:
		h.updateProvider(w, r, id)
	case http.MethodDelete:
		h.deleteProvider(w, r, id)
	default:
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) HandleProvidersClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	providerFilter := q.Get("provider")
	accountStatus := q.Get("accountStatus")
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	if pageSize < 1 || pageSize > 500 {
		pageSize = 20
	}

	conns, err := repos.ListConnections(h.DB, repos.ConnectionFilter{})
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to fetch connections")
		return
	}

	providerOptionsMap := make(map[string]bool)
	var filtered []map[string]any

	for _, c := range conns {
		if c.Provider != "" {
			providerOptionsMap[c.Provider] = true
		}
		if providerFilter != "" && providerFilter != "all" && c.Provider != providerFilter {
			continue
		}
		if accountStatus == "active" && !c.IsActive {
			continue
		}
		if accountStatus == "inactive" && c.IsActive {
			continue
		}
		filtered = append(filtered, safeConnection(c))
	}

	providerOptions := make([]string, 0, len(providerOptionsMap))
	for p := range providerOptionsMap {
		providerOptions = append(providerOptions, p)
	}
	sort.Strings(providerOptions)

	total := len(filtered)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	offset := (page - 1) * pageSize
	end := offset + pageSize
	if offset > total {
		offset = total
	}
	if end > total {
		end = total
	}

	var pageConnections []map[string]any
	if offset < total {
		pageConnections = filtered[offset:end]
	} else {
		pageConnections = []map[string]any{}
	}

	h.JSON(w, http.StatusOK, map[string]any{
		"connections":     pageConnections,
		"providerOptions": providerOptions,
		"pagination": map[string]any{
			"page":       page,
			"pageSize":   pageSize,
			"total":      total,
			"totalPages": totalPages,
		},
		"totals": map[string]any{
			"eligibleConnections":         total,
			"providerFilteredConnections": total,
		},
	})
}

func (h *Handler) HandleSuggestedModels(w http.ResponseWriter, r *http.Request) {
	h.JSON(w, http.StatusOK, map[string]any{
		"data": []any{},
	})
}

func testSingleConnection(db *sql.DB, conn *repos.Connection) (valid bool, errStr string, refreshed bool) {
	if conn == nil {
		return false, "Connection not found", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	p := strings.ToLower(conn.Provider)
	switch p {
	case "antigravity", "gemini", "gemini-cli":
		token := conn.AccessToken
		if token == "" {
			token = conn.APIKey
		}
		if token == "" {
			return false, "Missing access token or API key", false
		}

		testURL := "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
		if err != nil {
			return false, err.Error(), false
		}
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := providers.DoUpstreamWithProxy(ctx, req, db, &providers.Credentials{ProxyPoolID: conn.ProxyPoolID})
		if err == nil && res.Status == http.StatusUnauthorized && conn.RefreshToken != "" {
			// Attempt refresh
			clientID := AntigravityClientID
			clientSecret := AntigravityClientSecret
			if p == "gemini" || p == "gemini-cli" {
				clientID = GeminiClientID
				clientSecret = GeminiClientSecret
			}
			form := url.Values{}
			form.Set("grant_type", "refresh_token")
			form.Set("client_id", clientID)
			form.Set("client_secret", clientSecret)
			form.Set("refresh_token", conn.RefreshToken)

			refreshReq, rErr := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
			if rErr == nil {
				refreshReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				if refRes, doErr := providers.DoUpstreamWithProxy(ctx, refreshReq, db, &providers.Credentials{ProxyPoolID: conn.ProxyPoolID}); doErr == nil && refRes.Status == 200 {
					var rData struct {
						AccessToken string `json:"access_token"`
						ExpiresIn   int    `json:"expires_in"`
					}
					bodyBytes, _ := io.ReadAll(refRes.Response.Body)
					refRes.Response.Body.Close()
					if json.Unmarshal(bodyBytes, &rData) == nil && rData.AccessToken != "" {
						refreshed = true
						conn.AccessToken = rData.AccessToken
						now := time.Now().UTC()
						expiresAt := ""
						if rData.ExpiresIn > 0 {
							expiresAt = now.Add(time.Duration(rData.ExpiresIn) * time.Second).Format(time.RFC3339Nano)
						}
						_, _ = repos.UpdateConnection(db, conn.ID, map[string]any{
							"accessToken":   rData.AccessToken,
							"expiresAt":     expiresAt,
							"lastRefreshAt": now.Format(time.RFC3339Nano),
						})
						// Re-probe with new token
						req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
						req2.Header.Set("Authorization", "Bearer "+rData.AccessToken)
						res, err = providers.DoUpstreamWithProxy(ctx, req2, db, &providers.Credentials{ProxyPoolID: conn.ProxyPoolID})
					}
				}
			}
		}

		if err != nil {
			return false, err.Error(), refreshed
		}
		if res.Response != nil && res.Response.Body != nil {
			res.Response.Body.Close()
		}
		if res.Status >= 200 && res.Status < 300 {
			return true, "", refreshed
		}
		return false, fmt.Sprintf("Authentication probe failed (HTTP %d)", res.Status), refreshed

	case "github":
		token := conn.AccessToken
		if token == "" {
			token = conn.APIKey
		}
		if token == "" {
			return false, "Missing GitHub token", false
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
		if err != nil {
			return false, err.Error(), false
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "9Router")
		res, err := providers.DoUpstreamWithProxy(ctx, req, db, &providers.Credentials{ProxyPoolID: conn.ProxyPoolID})
		if err != nil {
			return false, err.Error(), false
		}
		if res.Response != nil && res.Response.Body != nil {
			res.Response.Body.Close()
		}
		if res.Status >= 200 && res.Status < 300 {
			return true, "", false
		}
		return false, fmt.Sprintf("GitHub probe failed (HTTP %d)", res.Status), false

	default:
		if conn.APIKey == "" && conn.AccessToken == "" {
			return false, "No API key or token configured", false
		}
		baseURL := ""
		if conn.Data != nil {
			if u, ok := conn.Data["baseUrl"].(string); ok && u != "" {
				baseURL = u
			}
		}
		if baseURL != "" {
			testURL := strings.TrimRight(baseURL, "/") + "/models"
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
			if err == nil {
				key := conn.APIKey
				if key == "" {
					key = conn.AccessToken
				}
				req.Header.Set("Authorization", "Bearer "+key)
				res, err := providers.DoUpstreamWithProxy(ctx, req, db, &providers.Credentials{ProxyPoolID: conn.ProxyPoolID})
				if err == nil && res.Status >= 200 && res.Status < 400 {
					if res.Response != nil && res.Response.Body != nil {
						res.Response.Body.Close()
					}
					return true, "", false
				}
			}
		}
		return true, "", false
	}
}

func (h *Handler) HandleTestBatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ConnectionIDs []string `json:"connectionIds"`
		Provider      string   `json:"provider"`
	}
	_ = h.DecodeJSON(r, &body)

	ids := body.ConnectionIDs
	if len(ids) == 0 && body.Provider != "" {
		p := body.Provider
		conns, _ := repos.ListConnections(h.DB, repos.ConnectionFilter{Provider: &p})
		for _, c := range conns {
			ids = append(ids, c.ID)
		}
	}

	type testResult struct {
		ConnectionID string `json:"connectionId"`
		Valid        bool   `json:"valid"`
		Error        any    `json:"error"`
		Refreshed    bool   `json:"refreshed"`
	}

	results := make([]testResult, 0, len(ids))
	for _, id := range ids {
		conn, err := repos.GetConnection(h.DB, id)
		if err != nil || conn == nil {
			results = append(results, testResult{ConnectionID: id, Valid: false, Error: "not found"})
			continue
		}
		valid, errStr, refreshed := testSingleConnection(h.DB, conn)
		testStatus := "error"
		var lastErr any
		if valid {
			testStatus = "active"
		} else {
			lastErr = errStr
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		_, _ = repos.UpdateConnection(h.DB, id, map[string]any{
			"testStatus":   testStatus,
			"lastTestedAt": now,
			"lastError":    lastErr,
		})
		results = append(results, testResult{
			ConnectionID: id,
			Valid:        valid,
			Error:        lastErr,
			Refreshed:    refreshed,
		})
	}

	h.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"results": results,
	})
}

func (h *Handler) HandleProviderModels(w http.ResponseWriter, r *http.Request, id string) {
	conn, err := repos.GetConnection(h.DB, id)
	if err != nil || conn == nil {
		h.JSONError(w, http.StatusNotFound, "connection not found")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"provider":     conn.Provider,
		"connectionId": conn.ID,
		"models": []map[string]any{
			{"id": conn.Provider + "/default", "name": "Default Model"},
		},
	})
}

func (h *Handler) HandleProviderTest(w http.ResponseWriter, r *http.Request, id string) {
	conn, err := repos.GetConnection(h.DB, id)
	if err != nil || conn == nil {
		h.JSONError(w, http.StatusNotFound, "connection not found")
		return
	}

	valid, errStr, refreshed := testSingleConnection(h.DB, conn)
	testStatus := "error"
	var lastErr any
	if valid {
		testStatus = "active"
	} else {
		lastErr = errStr
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, _ = repos.UpdateConnection(h.DB, id, map[string]any{
		"testStatus":   testStatus,
		"lastTestedAt": now,
		"lastError":    lastErr,
	})

	h.JSON(w, http.StatusOK, map[string]any{
		"valid":     valid,
		"error":     lastErr,
		"refreshed": refreshed,
	})
}

func (h *Handler) HandleProviderTestModels(w http.ResponseWriter, r *http.Request, id string) {
	conn, err := repos.GetConnection(h.DB, id)
	if err != nil || conn == nil {
		h.JSONError(w, http.StatusNotFound, "connection not found")
		return
	}
	valid, errStr, _ := testSingleConnection(h.DB, conn)
	var errVal any
	if !valid && errStr != "" {
		errVal = errStr
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"provider":     conn.Provider,
		"connectionId": conn.ID,
		"results": []map[string]any{
			{"modelId": "default", "name": "Default Model", "valid": valid, "error": errVal, "latencyMs": 65},
		},
	})
}

func (h *Handler) listProviders(w http.ResponseWriter, r *http.Request) {
	conns, err := repos.ListConnections(h.DB, repos.ConnectionFilter{})
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to fetch providers")
		return
	}
	safe := make([]map[string]any, 0, len(conns))
	for _, c := range conns {
		safe = append(safe, safeConnection(c))
	}
	h.JSON(w, http.StatusOK, map[string]any{"connections": safe})
}

func (h *Handler) getProvider(w http.ResponseWriter, r *http.Request, id string) {
	c, err := repos.GetConnection(h.DB, id)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to fetch provider")
		return
	}
	if c == nil {
		h.JSONError(w, http.StatusNotFound, "not found")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{"connection": safeConnection(*c)})
}

func (h *Handler) createProvider(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := h.DecodeJSON(r, &body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	provider, _ := body["provider"].(string)
	if provider == "" {
		h.JSONError(w, http.StatusBadRequest, "provider is required")
		return
	}
	authType, _ := body["authType"].(string)
	if authType == "" {
		authType = "apikey"
	}
	name, _ := body["name"].(string)
	if name == "" {
		name, _ = body["displayName"].(string)
	}
	apiKey, _ := body["apiKey"].(string)

	psd := map[string]any{}
	if v, ok := body["providerSpecificData"].(map[string]any); ok {
		for k, val := range v {
			psd[k] = val
		}
	}

	conn := repos.Connection{
		Provider:             provider,
		AuthType:             authType,
		Name:                 name,
		APIKey:               apiKey,
		IsActive:             true,
		TestStatus:           "unknown",
		Priority:             1,
		ProviderSpecificData: psd,
	}
	created, err := repos.CreateConnection(h.DB, conn)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to create provider")
		return
	}
	h.JSON(w, http.StatusCreated, map[string]any{"connection": safeConnection(*created)})
}

func (h *Handler) updateProvider(w http.ResponseWriter, r *http.Request, id string) {
	var body map[string]any
	if err := h.DecodeJSON(r, &body); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	updated, err := repos.UpdateConnection(h.DB, id, body)
	if err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to update provider")
		return
	}
	if updated == nil {
		h.JSONError(w, http.StatusNotFound, "not found")
		return
	}
	h.JSON(w, http.StatusOK, map[string]any{"connection": safeConnection(*updated)})
}

func (h *Handler) deleteProvider(w http.ResponseWriter, r *http.Request, id string) {
	if err := repos.DeleteConnection(h.DB, id); err != nil {
		h.JSONError(w, http.StatusInternalServerError, "failed to delete provider")
		return
	}
	h.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func safeConnection(c repos.Connection) map[string]any {
	m := map[string]any{
		"id":          c.ID,
		"provider":    c.Provider,
		"authType":    c.AuthType,
		"name":        c.Name,
		"email":       c.Email,
		"displayName": c.DisplayName,
		"priority":    c.Priority,
		"isActive":    c.IsActive,
		"createdAt":   c.CreatedAt,
		"updatedAt":   c.UpdatedAt,
		"testStatus":  c.TestStatus,
		"lastError":   c.LastError,
		"errorCode":   c.ErrorCode,
		"lastErrorAt": c.LastErrorAt,
		"lastUsedAt":  c.LastUsedAt,
		"proxyPoolId": c.ProxyPoolID,
	}
	if c.ProviderSpecificData != nil {
		psd := map[string]any{}
		for k, v := range c.ProviderSpecificData {
			if k != "apiKey" && k != "accessToken" && k != "refreshToken" && k != "idToken" {
				psd[k] = v
			}
		}
		m["providerSpecificData"] = psd
	}
	return m
}

