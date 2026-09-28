package providers

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

	"github.com/dresar/go-9router/internal/logging"
)

type ProviderAdapter interface {
	ID() string
	BuildRequest(ctx context.Context, body map[string]any, creds *Credentials) (*http.Request, error)
	ParseError(resp *http.Response) (string, bool)
}

const DefaultClientTimeout = 30 * time.Second

type UpstreamResult struct {
	Response *http.Response
	Error    string
	Status   int
	Success  bool
}

var httpClient = &http.Client{
	Timeout: 0,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

func DoUpstream(ctx context.Context, req *http.Request) (*UpstreamResult, error) {
	req = req.WithContext(ctx)
	resp, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return &UpstreamResult{Error: "request cancelled", Status: 499, Success: false}, nil
		}
		return nil, fmt.Errorf("upstream request: %w", err)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return &UpstreamResult{
			Response: resp,
			Error:    string(body),
			Status:   resp.StatusCode,
			Success:  false,
		}, nil
	}
	return &UpstreamResult{Response: resp, Status: resp.StatusCode, Success: true}, nil
}

func NewJSONRequest(ctx context.Context, method, rawURL string, body any, headers map[string]string) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

func ValidateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must be http or https")
	}
	host := strings.ToLower(u.Hostname())
	blocked := []string{"169.254.", "10.", "172.16.", "192.168.", "127.", "::1", "0.0.0.0", "metadata."}
	for _, prefix := range blocked {
		if strings.HasPrefix(host, prefix) {
			return fmt.Errorf("URL host is not allowed")
		}
	}
	return nil
}

func ShouldFallback(status int) bool {
	switch status {
	case 429, 503, 529, 402:
		return true
	case 401, 403, 404:
		return true
	}
	if status >= 500 {
		return true
	}
	return false
}

func LogRequest(provider, model, connectionName string, streaming bool) {
	mode := "json"
	if streaming {
		mode = "sse"
	}
	logging.Info("CHAT", fmt.Sprintf("▶ %s/%s acc:%s [%s]", provider, model, connectionName, mode))
}
