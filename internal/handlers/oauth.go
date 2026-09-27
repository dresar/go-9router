package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dresar/go-9router/internal/storage/repos"
	"github.com/google/uuid"
)

const (
	AntigravityClientID     = "1071006060591-tmhssin2h21lcre235vtolojh4g403ep.apps.googleusercontent.com"
	AntigravityClientSecret = "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"
	GeminiClientID          = "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com"
	GeminiClientSecret      = "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"
)

func (h *Handler) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	trimmed := strings.TrimPrefix(path, "/api/oauth/")
	trimmed = strings.Trim(trimmed, "/")
	parts := strings.Split(trimmed, "/")

	if len(parts) == 0 || parts[0] == "" {
		h.JSON(w, http.StatusOK, map[string]any{"status": "ok"})
		return
	}

	provider := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "authorize":
		h.handleOAuthAuthorize(w, r, provider)
	case "exchange":
		h.handleOAuthExchange(w, r, provider)
	case "device-code":
		h.handleOAuthDeviceCode(w, r, provider)
	case "poll":
		h.handleOAuthPoll(w, r, provider)
	case "ide-status":
		h.JSON(w, http.StatusOK, map[string]any{"installed": false, "path": nil})
	case "start-proxy":
		h.JSON(w, http.StatusOK, map[string]any{"success": false, "reason": "server_side_proxy_disabled"})
	case "stop-proxy":
		h.JSON(w, http.StatusOK, map[string]any{"success": true})
	case "poll-status":
		h.JSON(w, http.StatusOK, map[string]any{"status": "idle"})
	case "auto-import":
		h.JSON(w, http.StatusOK, map[string]any{"found": false, "error": "No local session detected"})
	default:
		if r.Method == http.MethodGet {
			h.handleOAuthAuthorize(w, r, provider)
			return
		}
		h.JSON(w, http.StatusOK, map[string]any{"success": true})
	}
}

func (h *Handler) handleOAuthAuthorize(w http.ResponseWriter, r *http.Request, provider string) {
	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = "http://localhost:20127/callback"
	}
	state := q.Get("state")
	if state == "" {
		state = uuid.NewString()
	}

	switch provider {
	case "antigravity":
		scopes := []string{
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/cclog",
			"https://www.googleapis.com/auth/experimentsandconfigs",
		}
		params := url.Values{}
		params.Set("client_id", AntigravityClientID)
		params.Set("response_type", "code")
		params.Set("redirect_uri", redirectURI)
		params.Set("scope", strings.Join(scopes, " "))
		params.Set("state", state)
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")

		authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
		h.JSON(w, http.StatusOK, map[string]any{
			"authUrl":     authURL,
			"state":       state,
			"flowType":    "authorization_code",
			"redirectUri": redirectURI,
		})

	case "gemini", "gemini-cli":
		scopes := []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/generative-language",
		}
		params := url.Values{}
		params.Set("client_id", GeminiClientID)
		params.Set("response_type", "code")
		params.Set("redirect_uri", redirectURI)
		params.Set("scope", strings.Join(scopes, " "))
		params.Set("state", state)
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")

		authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
		h.JSON(w, http.StatusOK, map[string]any{
			"authUrl":     authURL,
			"state":       state,
			"flowType":    "authorization_code",
			"redirectUri": redirectURI,
		})

	case "claude":
		params := url.Values{}
		params.Set("client_id", "9d1c250a-e54e-4509-98cf-1934be9440e5")
		params.Set("response_type", "code")
		params.Set("redirect_uri", redirectURI)
		params.Set("scope", "user:read")
		params.Set("state", state)
		authURL := "https://claude.ai/oauth/authorize?" + params.Encode()
		h.JSON(w, http.StatusOK, map[string]any{
			"authUrl":     authURL,
			"state":       state,
			"flowType":    "authorization_code",
			"redirectUri": redirectURI,
		})

	case "codex":
		params := url.Values{}
		params.Set("client_id", "app-p4M4s8B4h8o7s5")
		params.Set("response_type", "code")
		params.Set("redirect_uri", redirectURI)
		params.Set("scope", "openid profile email offline_access")
		params.Set("state", state)
		authURL := "https://auth.openai.com/authorize?" + params.Encode()
		h.JSON(w, http.StatusOK, map[string]any{
			"authUrl":     authURL,
			"state":       state,
			"flowType":    "authorization_code",
			"redirectUri": redirectURI,
		})

	default:
		// Fallback Google OAuth client for other Google-based tools
		params := url.Values{}
		params.Set("client_id", AntigravityClientID)
		params.Set("response_type", "code")
		params.Set("redirect_uri", redirectURI)
		params.Set("state", state)
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
		authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
		h.JSON(w, http.StatusOK, map[string]any{
			"authUrl":     authURL,
			"state":       state,
			"flowType":    "authorization_code",
			"redirectUri": redirectURI,
		})
	}
}

type exchangeRequest struct {
	Code         string `json:"code"`
	RedirectURI  string `json:"redirectUri"`
	CodeVerifier string `json:"codeVerifier"`
	State        string `json:"state"`
}

func (h *Handler) handleOAuthExchange(w http.ResponseWriter, r *http.Request, provider string) {
	if r.Method != http.MethodPost {
		h.JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req exchangeRequest
	if err := h.DecodeJSON(r, &req); err != nil {
		h.JSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		h.JSONError(w, http.StatusBadRequest, "missing authorization code")
		return
	}
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = "http://localhost:20127/callback"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	switch provider {
	case "antigravity", "gemini", "gemini-cli":
		clientID := AntigravityClientID
		clientSecret := AntigravityClientSecret
		if provider == "gemini" || provider == "gemini-cli" {
			clientID = GeminiClientID
			clientSecret = GeminiClientSecret
		}

		form := url.Values{}
		form.Set("grant_type", "authorization_code")
		form.Set("client_id", clientID)
		form.Set("client_secret", clientSecret)
		form.Set("code", code)
		form.Set("redirect_uri", redirectURI)

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, "failed to prepare token request")
			return
		}
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpReq.Header.Set("Accept", "application/json")

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			h.JSONError(w, http.StatusBadGateway, fmt.Sprintf("failed to contact token endpoint: %v", err))
			return
		}
		defer resp.Body.Close()

		respBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			h.JSONError(w, http.StatusBadRequest, fmt.Sprintf("token exchange error (%d): %s", resp.StatusCode, string(respBytes)))
			return
		}

		var tokenResp struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
			IDToken      string `json:"id_token"`
			TokenType    string `json:"token_type"`
		}
		if err := json.Unmarshal(respBytes, &tokenResp); err != nil {
			h.JSONError(w, http.StatusInternalServerError, "invalid token response")
			return
		}

		// Fetch User Info
		userEmail := ""
		userName := ""
		userInfoReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v1/userinfo?alt=json", nil)
		if userInfoReq != nil {
			userInfoReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
			if uResp, uErr := http.DefaultClient.Do(userInfoReq); uErr == nil {
				defer uResp.Body.Close()
				var uInfo struct {
					Email string `json:"email"`
					Name  string `json:"name"`
				}
				_ = json.NewDecoder(uResp.Body).Decode(&uInfo)
				userEmail = uInfo.Email
				userName = uInfo.Name
			}
		}

		// Fetch Code Assist Project ID for Antigravity
		projectID := ""
		if provider == "antigravity" {
			loadReqBody := []byte(`{"metadata":{"ideType":9,"platform":5,"pluginType":2}}`)
			loadReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist", bytes.NewReader(loadReqBody))
			if loadReq != nil {
				loadReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
				loadReq.Header.Set("Content-Type", "application/json")
				loadReq.Header.Set("User-Agent", "antigravity/ide/2.11.0 windows/amd64")
				if lResp, lErr := http.DefaultClient.Do(loadReq); lErr == nil {
					defer lResp.Body.Close()
					var lData struct {
						CompanionProject any `json:"cloudaicompanionProject"`
					}
					_ = json.NewDecoder(lResp.Body).Decode(&lData)
					if m, ok := lData.CompanionProject.(map[string]any); ok {
						if pid, ok2 := m["id"].(string); ok2 {
							projectID = pid
						}
					} else if s, ok := lData.CompanionProject.(string); ok {
						projectID = s
					}
				}
			}
		}

		now := time.Now().UTC()
		expiresAt := ""
		if tokenResp.ExpiresIn > 0 {
			expiresAt = now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Format(time.RFC3339Nano)
		}

		connID := fmt.Sprintf("p_%d_%s", now.UnixMilli(), strings.ToLower(uuid.NewString()[:6]))
		conn := repos.Connection{
			ID:           connID,
			Provider:     provider,
			AuthType:     "oauth",
			Name:         userName,
			Email:        userEmail,
			Priority:     1,
			IsActive:     true,
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
			IDToken:      tokenResp.IDToken,
			ExpiresIn:    tokenResp.ExpiresIn,
			ExpiresAt:    expiresAt,
			ProjectID:    projectID,
			TestStatus:   "active",
			CreatedAt:    now.Format(time.RFC3339Nano),
			UpdatedAt:    now.Format(time.RFC3339Nano),
		}

		created, err := repos.CreateConnection(h.DB, conn)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save connection: %v", err))
			return
		}

		h.JSON(w, http.StatusOK, map[string]any{
			"success":    true,
			"connection": safeConnection(*created),
		})

	default:
		now := time.Now().UTC()
		connID := fmt.Sprintf("p_%d_%s", now.UnixMilli(), strings.ToLower(uuid.NewString()[:6]))
		conn := repos.Connection{
			ID:          connID,
			Provider:    provider,
			AuthType:    "oauth",
			AccessToken: code,
			IsActive:    true,
			TestStatus:  "active",
			CreatedAt:   now.Format(time.RFC3339Nano),
			UpdatedAt:   now.Format(time.RFC3339Nano),
		}
		created, err := repos.CreateConnection(h.DB, conn)
		if err != nil {
			h.JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save connection: %v", err))
			return
		}
		h.JSON(w, http.StatusOK, map[string]any{
			"success":    true,
			"connection": safeConnection(*created),
		})
	}
}

func (h *Handler) handleOAuthDeviceCode(w http.ResponseWriter, r *http.Request, provider string) {
	deviceCode := uuid.NewString()
	userCode := strings.ToUpper(uuid.NewString()[:8])
	verifyURL := "https://github.com/login/device"
	if provider == "qoder" || provider == "qoder-cn" {
		verifyURL = "https://qoder.sh/activate"
	}
	h.JSON(w, http.StatusOK, map[string]any{
		"device_code":               deviceCode,
		"user_code":                 userCode,
		"verification_uri":          verifyURL,
		"verification_uri_complete": fmt.Sprintf("%s?user_code=%s", verifyURL, userCode),
		"expires_in":                300,
		"interval":                  5,
	})
}

func (h *Handler) handleOAuthPoll(w http.ResponseWriter, r *http.Request, provider string) {
	h.JSON(w, http.StatusOK, map[string]any{
		"error": "authorization_pending",
	})
}
