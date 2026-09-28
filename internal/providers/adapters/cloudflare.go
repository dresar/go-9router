package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dresar/go-9router/internal/providers"
)

type CloudflareAIAdapter struct{}

func (CloudflareAIAdapter) ID() string { return "cloudflare-ai" }

func (CloudflareAIAdapter) BuildRequest(ctx context.Context, body map[string]any, creds *providers.Credentials) (*http.Request, error) {
	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = creds.AccessToken
	}

	accountId := ""
	if creds.ProviderSpecificData != nil {
		if a, ok := creds.ProviderSpecificData["accountId"].(string); ok && a != "" {
			accountId = a
		} else if a, ok := creds.ProviderSpecificData["account_id"].(string); ok && a != "" {
			accountId = a
		}
	}
	if accountId == "" && creds.ProjectID != "" {
		accountId = creds.ProjectID
	}
	if accountId == "" {
		return nil, fmt.Errorf("cloudflare-ai requires accountId in credentials")
	}

	// Normalize model ID in body so it matches Cloudflare's @cf/ namespace
	if m, ok := body["model"].(string); ok {
		clean := m
		if strings.HasPrefix(clean, "cloudflare-ai/") {
			clean = strings.TrimPrefix(clean, "cloudflare-ai/")
		} else if strings.HasPrefix(clean, "cf/") {
			clean = strings.TrimPrefix(clean, "cf/")
		}
		if !strings.HasPrefix(clean, "@cf/") && !strings.Contains(clean, "/") {
			clean = "@cf/" + clean
		}
		body["model"] = clean
	}

	targetURL := creds.BaseURL
	if targetURL == "" {
		targetURL = fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/v1/chat/completions", accountId)
	} else if strings.Contains(targetURL, "{accountId}") {
		targetURL = strings.ReplaceAll(targetURL, "{accountId}", accountId)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
	}

	return providers.NewJSONRequest(ctx, http.MethodPost, targetURL, body, headers)
}

func (CloudflareAIAdapter) ParseError(resp *http.Response) (string, bool) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	resp.Body = io.NopCloser(bReader(body))

	var cfErr struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &cfErr) == nil {
		if len(cfErr.Errors) > 0 && cfErr.Errors[0].Message != "" {
			return cfErr.Errors[0].Message, true
		}
		if cfErr.Error.Message != "" {
			return cfErr.Error.Message, true
		}
	}

	return string(body), true
}
