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

type Anthropic struct{}

func (Anthropic) ID() string { return "anthropic" }

func (Anthropic) BuildRequest(ctx context.Context, body map[string]any, creds *providers.Credentials) (*http.Request, error) {
	baseURL := "https://api.anthropic.com"
	if creds.ProviderSpecificData != nil {
		if u, ok := creds.ProviderSpecificData["baseUrl"].(string); ok && u != "" {
			baseURL = u
		}
	}

	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = creds.AccessToken
	}

	baseURL = strings.TrimRight(baseURL, "/")
	targetURL := baseURL
	if !strings.Contains(targetURL, "/messages") {
		if strings.HasSuffix(targetURL, "/v1") {
			targetURL += "/messages"
		} else {
			targetURL += "/v1/messages"
		}
	}

	req, err := providers.NewJSONRequest(ctx, http.MethodPost, targetURL, body, map[string]string{
		"x-api-key":         apiKey,
		"anthropic-version": "2023-06-01",
	})
	if err != nil {
		return nil, fmt.Errorf("anthropic build request: %w", err)
	}
	return req, nil
}

func (Anthropic) ParseError(resp *http.Response) (string, bool) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	resp.Body = io.NopCloser(bReader(body))
	var e struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error.Message != "" {
		return e.Error.Message, true
	}
	return string(body), true
}
