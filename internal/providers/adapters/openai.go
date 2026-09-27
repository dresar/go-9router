package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dresar/go-9router/internal/providers"
)

type OpenAI struct{}

func (OpenAI) ID() string { return "openai" }

func (OpenAI) BaseURL() string { return "https://api.openai.com" }

func (OpenAI) BuildRequest(ctx context.Context, body map[string]any, creds *providers.Credentials) (*http.Request, error) {
	baseURL := "https://api.openai.com"
	if creds.ProviderSpecificData != nil {
		if u, ok := creds.ProviderSpecificData["baseUrl"].(string); ok && u != "" {
			baseURL = u
		}
	}

	model, _ := body["model"].(string)
	endpoint := chatEndpoint(baseURL, model, body)

	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = creds.AccessToken
	}

	req, err := providers.NewJSONRequest(ctx, http.MethodPost, endpoint, body, map[string]string{
		"Authorization": "Bearer " + apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("openai build request: %w", err)
	}

	if pref, ok := creds.ProviderSpecificData["prefix"].(string); ok && pref != "" {
		req.Header.Set("Authorization", "Bearer "+pref+apiKey)
	}

	return req, nil
}

func (OpenAI) ParseError(resp *http.Response) (string, bool) {
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

func chatEndpoint(baseURL, model string, body map[string]any) string {
	if isEmbeddingRequest(body) {
		return baseURL + "/v1/embeddings"
	}
	if isImageRequest(body) {
		return baseURL + "/v1/images/generations"
	}
	return baseURL + "/v1/chat/completions"
}

func isEmbeddingRequest(body map[string]any) bool {
	_, hasInput := body["input"]
	_, hasModel := body["model"]
	_, hasMessages := body["messages"]
	return hasInput && hasModel && !hasMessages
}

func isImageRequest(body map[string]any) bool {
	_, hasPrompt := body["prompt"]
	_, hasN := body["n"]
	_, hasMessages := body["messages"]
	return hasPrompt && hasN && !hasMessages
}

func bReader(b []byte) *byteReader {
	return &byteReader{b: b, pos: 0}
}

type byteReader struct {
	b   []byte
	pos int
}

func (br *byteReader) Read(p []byte) (int, error) {
	if br.pos >= len(br.b) {
		return 0, io.EOF
	}
	n := copy(p, br.b[br.pos:])
	br.pos += n
	return n, nil
}
