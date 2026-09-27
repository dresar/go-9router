package adapters

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/dresar/go-9router/internal/providers"
)

type GenericAPIKey struct {
	id      string
	baseURL string
	authHeader string
}

func NewGenericAPIKey(id, baseURL, authHeader string) *GenericAPIKey {
	return &GenericAPIKey{id: id, baseURL: baseURL, authHeader: authHeader}
}

func (g *GenericAPIKey) ID() string { return g.id }

func (g *GenericAPIKey) BuildRequest(ctx context.Context, body map[string]any, creds *providers.Credentials) (*http.Request, error) {
	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = creds.AccessToken
	}
	headers := map[string]string{}
	if g.authHeader == "x-api-key" {
		headers["x-api-key"] = apiKey
	} else {
		headers["Authorization"] = "Bearer " + apiKey
	}
	req, err := providers.NewJSONRequest(ctx, http.MethodPost, g.baseURL+"/v1/chat/completions", body, headers)
	if err != nil {
		return nil, fmt.Errorf("%s build request: %w", g.id, err)
	}
	return req, nil
}

func (g *GenericAPIKey) ParseError(resp *http.Response) (string, bool) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	resp.Body = io.NopCloser(bReader(body))
	return string(body), true
}

var KnownProviders = map[string]providers.ProviderAdapter{
	"openai":    OpenAI{},
	"anthropic": Anthropic{},
	"claude":    Anthropic{},
	"groq":      NewGenericAPIKey("groq", "https://api.groq.com/openai", "Bearer"),
	"openrouter": NewGenericAPIKey("openrouter", "https://openrouter.ai/api", "Bearer"),
	"deepseek":  NewGenericAPIKey("deepseek", "https://api.deepseek.com", "Bearer"),
	"mistral":   NewGenericAPIKey("mistral", "https://api.mistral.ai", "Bearer"),
	"cohere":    NewGenericAPIKey("cohere", "https://api.cohere.com", "Bearer"),
	"perplexity": NewGenericAPIKey("perplexity", "https://api.perplexity.ai", "Bearer"),
	"together":  NewGenericAPIKey("together", "https://api.together.xyz", "Bearer"),
	"fireworks": NewGenericAPIKey("fireworks", "https://api.fireworks.ai/inference", "Bearer"),
	"huggingface": NewGenericAPIKey("huggingface", "https://api-inference.huggingface.co", "Bearer"),
	"nvidia":    NewGenericAPIKey("nvidia", "https://integrate.api.nvidia.com", "Bearer"),
	"siliconflow": NewGenericAPIKey("siliconflow", "https://api.siliconflow.cn", "Bearer"),
	"featherless": NewGenericAPIKey("featherless", "https://api.featherless.ai", "Bearer"),
	"nebius":    NewGenericAPIKey("nebius", "https://api.studio.nebius.ai", "Bearer"),
	"sambanova": NewGenericAPIKey("sambanova", "https://api.sambanova.ai", "Bearer"),
	"xai":       NewGenericAPIKey("xai", "https://api.x.ai", "Bearer"),
	"venice":    NewGenericAPIKey("venice", "https://api.venice.ai", "Bearer"),
	"hyperbolic": NewGenericAPIKey("hyperbolic", "https://api.hyperbolic.xyz", "Bearer"),
}

func GetAdapter(providerID string) (providers.ProviderAdapter, bool) {
	a, ok := KnownProviders[providerID]
	return a, ok
}
