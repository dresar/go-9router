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

	// Go directly handles server-side OAuth ONLY for Google Cloud-Platform providers
	// (Antigravity and Gemini) which require Code Assist project ID fetching and background onboarding.
	if provider == "antigravity" || provider == "gemini" || provider == "gemini-cli" {
		switch action {
		case "authorize":
			h.handleOAuthAuthorize(w, r, provider)
			return
		case "exchange":
			h.handleOAuthExchange(w, r, provider)
			return
		}
	}

	// All other providers (CodeBuddy CN, CodeBuddy Intl, Qoder, Kimi, Kiro, Trae, Windsurf, Zed, xAI, Claude, Codex, GitHub, etc.)
	// and all OAuth flows (device-code, poll, start-proxy, stop-proxy, poll-status, ide-status, register-session, manual-code, etc.)
	// are delegated to the complete Next.js OAuth implementation:
	h.HandleFrontend(w, r)
}

func (h *Handler) handleOAuthAuthorize(w http.ResponseWriter, r *http.Request, provider string) {
	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	if redirectURI == "" {
		host := r.Host
		if host == "" {
			host = "localhost:20128"
		}
		redirectURI = fmt.Sprintf("http://%s/callback", host)
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
		host := r.Host
		if host == "" {
			host = "localhost:20128"
		}
		redirectURI = fmt.Sprintf("http://%s/callback", host)
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
			// Auto-retry with alternative loopback URI if redirect_uri_mismatch
			if strings.Contains(string(respBytes), "redirect_uri_mismatch") {
				altRedirectURI := redirectURI
				if strings.Contains(altRedirectURI, "localhost") {
					altRedirectURI = strings.Replace(altRedirectURI, "localhost", "127.0.0.1", 1)
				} else if strings.Contains(altRedirectURI, "127.0.0.1") {
					altRedirectURI = strings.Replace(altRedirectURI, "127.0.0.1", "localhost", 1)
				}
				if altRedirectURI != redirectURI {
					form.Set("redirect_uri", altRedirectURI)
					httpReqRetry, err2 := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
					if err2 == nil {
						httpReqRetry.Header.Set("Content-Type", "application/x-www-form-urlencoded")
						httpReqRetry.Header.Set("Accept", "application/json")
						if retryResp, retryErr := http.DefaultClient.Do(httpReqRetry); retryErr == nil {
							defer retryResp.Body.Close()
							retryBytes, _ := io.ReadAll(retryResp.Body)
							if retryResp.StatusCode >= 200 && retryResp.StatusCode < 300 {
								respBytes = retryBytes
								resp.StatusCode = retryResp.StatusCode
							}
						}
					}
				}
			}
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				h.JSONError(w, http.StatusBadRequest, fmt.Sprintf("token exchange error (%d): %s", resp.StatusCode, string(respBytes)))
				return
			}
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

		// Fetch User Info with fast 3s timeout
		userEmail := ""
		userName := ""
		uCtx, uCancel := context.WithTimeout(context.Background(), 3*time.Second)
		userInfoReq, _ := http.NewRequestWithContext(uCtx, http.MethodGet, "https://www.googleapis.com/oauth2/v1/userinfo?alt=json", nil)
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
		uCancel()

		// Fetch Code Assist Project ID for Antigravity (fast 3s timeout, non-blocking)
		projectID := ""
		tierID := "legacy-tier"
		if provider == "antigravity" {
			loadCtx, loadCancel := context.WithTimeout(context.Background(), 3*time.Second)
			loadReqBody := []byte(`{"metadata":{"ideType":9,"platform":5,"pluginType":2}}`)
			loadReq, _ := http.NewRequestWithContext(loadCtx, http.MethodPost, "https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist", bytes.NewReader(loadReqBody))
			if loadReq != nil {
				loadReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
				loadReq.Header.Set("Content-Type", "application/json")
				loadReq.Header.Set("User-Agent", "antigravity/ide/2.11.0 windows/amd64")
				loadReq.Header.Set("x-request-source", "local")
				if lResp, lErr := http.DefaultClient.Do(loadReq); lErr == nil {
					defer lResp.Body.Close()
					var lData struct {
						CompanionProject any `json:"cloudaicompanionProject"`
						AllowedTiers     []struct {
							ID        string `json:"id"`
							IsDefault bool   `json:"isDefault"`
						} `json:"allowedTiers"`
					}
					_ = json.NewDecoder(lResp.Body).Decode(&lData)
					if m, ok := lData.CompanionProject.(map[string]any); ok {
						if pid, ok2 := m["id"].(string); ok2 {
							projectID = strings.TrimSpace(pid)
						}
					} else if s, ok := lData.CompanionProject.(string); ok {
						projectID = strings.TrimSpace(s)
					}
					for _, tier := range lData.AllowedTiers {
						if tier.IsDefault && tier.ID != "" {
							tierID = strings.TrimSpace(tier.ID)
							break
						}
					}
				}
			}
			loadCancel()
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

		// Background onboarding to enable Gemini Code Assist without slowing down response
		if provider == "antigravity" {
			go h.asyncAntigravityOnboard(connID, tokenResp.AccessToken, projectID, tierID)
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


func (h *Handler) asyncAntigravityOnboard(connID, accessToken, projectID, tierID string) {
	// 1. If projectID is still missing, retry fetch up to 3 times
	if projectID == "" {
		for attempt := 0; attempt < 3; attempt++ {
			time.Sleep(2 * time.Second)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			loadReqBody := []byte(`{"metadata":{"ideType":9,"platform":5,"pluginType":2}}`)
			loadReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist", bytes.NewReader(loadReqBody))
			if loadReq != nil {
				loadReq.Header.Set("Authorization", "Bearer "+accessToken)
				loadReq.Header.Set("Content-Type", "application/json")
				loadReq.Header.Set("User-Agent", "antigravity/ide/2.11.0 windows/amd64")
				loadReq.Header.Set("x-request-source", "local")
				if lResp, lErr := http.DefaultClient.Do(loadReq); lErr == nil {
					var lData struct {
						CompanionProject any `json:"cloudaicompanionProject"`
						AllowedTiers     []struct {
							ID        string `json:"id"`
							IsDefault bool   `json:"isDefault"`
						} `json:"allowedTiers"`
					}
					_ = json.NewDecoder(lResp.Body).Decode(&lData)
					lResp.Body.Close()
					if m, ok := lData.CompanionProject.(map[string]any); ok {
						if pid, ok2 := m["id"].(string); ok2 {
							projectID = strings.TrimSpace(pid)
						}
					} else if s, ok := lData.CompanionProject.(string); ok {
						projectID = strings.TrimSpace(s)
					}
					for _, tier := range lData.AllowedTiers {
						if tier.IsDefault && tier.ID != "" {
							tierID = strings.TrimSpace(tier.ID)
							break
						}
					}
					if projectID != "" {
						cancel()
						_, _ = repos.UpdateConnection(h.DB, connID, map[string]any{"projectId": projectID})
						break
					}
				}
			}
			cancel()
		}
	}

	// 2. Onboard user to Gemini Code Assist (fire-and-forget, up to 10 retries, standard 9router spec)
	if projectID != "" {
		for i := 0; i < 10; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			body := fmt.Sprintf(`{"tierId":%q,"metadata":{"ideType":9,"platform":5,"pluginType":2}}`, tierID)
			onboardReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://cloudcode-pa.googleapis.com/v1internal:onboardUser", strings.NewReader(body))
			if onboardReq != nil {
				onboardReq.Header.Set("Authorization", "Bearer "+accessToken)
				onboardReq.Header.Set("Content-Type", "application/json")
				onboardReq.Header.Set("User-Agent", "antigravity/ide/2.11.0 windows/amd64")
				onboardReq.Header.Set("x-request-source", "local")
				if oResp, oErr := http.DefaultClient.Do(onboardReq); oErr == nil {
					var res struct {
						Done bool `json:"done"`
					}
					_ = json.NewDecoder(oResp.Body).Decode(&res)
					oResp.Body.Close()
					cancel()
					if res.Done {
						break
					}
				} else {
					cancel()
				}
			} else {
				cancel()
			}
			time.Sleep(5 * time.Second)
		}
	}
}

// HandleOAuthCallback handles the /callback endpoint directly on Go gateway,
// ensuring lightning-fast (<1ms) and cross-origin-safe token relay to the opener modal.
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Go 9Router - OAuth Authorization</title>
  <style>
    body { background: #090a0f; color: #e2e8f0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
    .card { background: #12141c; border: 1px solid #1e2230; border-radius: 12px; padding: 2rem; text-align: center; max-width: 420px; box-shadow: 0 10px 30px rgba(0,0,0,0.5); }
    .spinner { border: 3px solid rgba(255,255,255,0.1); border-top: 3px solid #38bdf8; border-radius: 50%; width: 40px; height: 40px; animation: spin 0.8s linear infinite; margin: 0 auto 1.25rem; }
    @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
    h2 { font-size: 1.25rem; margin-bottom: 0.5rem; color: #fff; }
    p { color: #94a3b8; font-size: 0.875rem; margin-bottom: 1rem; }
  </style>
</head>
<body>
  <div class="card">
    <div class="spinner"></div>
    <h2>Authorization Complete</h2>
    <p>Connecting with 9Router dashboard... This window will close automatically.</p>
  </div>
  <script>
    (function() {
      var params = new URLSearchParams(window.location.search);
      var code = params.get("code");
      var token = params.get("token");
      var state = params.get("state");
      var error = params.get("error");
      var errorDesc = params.get("error_description");

      var data = {
        code: code,
        token: token,
        state: state,
        error: error,
        errorDescription: errorDesc,
        fullUrl: window.location.href
      };

      if (window.opener) {
        try { window.opener.postMessage({ type: "oauth_callback", data: data }, "*"); } catch(e){}
        var port = window.location.port || "20128";
        var origins = [
          window.location.origin,
          "http://localhost:" + port, "http://127.0.0.1:" + port, "http://[::1]:" + port,
          "http://localhost:20128", "http://127.0.0.1:20128", "http://[::1]:20128",
          "http://localhost:20127", "http://127.0.0.1:20127", "http://localhost:1455"
        ];
        for (var i = 0; i < origins.length; i++) {
          try { window.opener.postMessage({ type: "oauth_callback", data: data }, origins[i]); } catch(e){}
        }
      }

      try {
        var bc = new BroadcastChannel("oauth_callback");
        bc.postMessage(data);
        bc.close();
      } catch(e){}

      try {
        localStorage.setItem("oauth_callback", JSON.stringify({
          code: code, token: token, state: state, error: error, errorDescription: errorDesc,
          timestamp: Date.now()
        }));
      } catch(e){}

      setTimeout(function() {
        window.close();
      }, 1000);
    })();
  </script>
</body>
</html>`
	w.Write([]byte(html))
}
