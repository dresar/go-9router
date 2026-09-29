package adapters

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dresar/go-9router/internal/providers"
)

type GenericAPIKey struct {
	id         string
	baseURL    string
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
	targetURL := g.baseURL
	if !strings.Contains(targetURL, "/chat/completions") {
		if strings.Contains(targetURL, "/v1") {
			targetURL = strings.TrimRight(targetURL, "/") + "/chat/completions"
		} else if strings.HasSuffix(targetURL, "/") {
			targetURL += "v1/chat/completions"
		} else {
			targetURL += "/v1/chat/completions"
		}
	}
	req, err := providers.NewJSONRequest(ctx, http.MethodPost, targetURL, body, headers)
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

type GeminiAdapter struct{}

func (g GeminiAdapter) ID() string { return "gemini" }

func (g GeminiAdapter) BuildRequest(ctx context.Context, body map[string]any, creds *providers.Credentials) (*http.Request, error) {
	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = creds.AccessToken
	}
	headers := map[string]string{
		"Authorization":  "Bearer " + apiKey,
		"x-goog-api-key": apiKey,
	}
	return providers.NewJSONRequest(ctx, http.MethodPost, "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", body, headers)
}

func (g GeminiAdapter) ParseError(resp *http.Response) (string, bool) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	resp.Body = io.NopCloser(bReader(body))
	return string(body), true
}

type GitHubCopilotAdapter struct{}

func (g GitHubCopilotAdapter) ID() string { return "github" }

func (g GitHubCopilotAdapter) BuildRequest(ctx context.Context, body map[string]any, creds *providers.Credentials) (*http.Request, error) {
	token := creds.AccessToken
	if token == "" {
		token = creds.APIKey
	}
	headers := map[string]string{
		"Authorization":          "Bearer " + token,
		"copilot-integration-id": "vscode-chat",
		"editor-version":         "vscode/1.110.0",
		"editor-plugin-version":  "copilot-chat/0.38.0",
		"user-agent":             "GitHubCopilotChat/0.38.0",
		"openai-intent":          "conversation-panel",
	}
	return providers.NewJSONRequest(ctx, http.MethodPost, "https://api.githubcopilot.com/chat/completions", body, headers)
}

func (g GitHubCopilotAdapter) ParseError(resp *http.Response) (string, bool) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	resp.Body = io.NopCloser(bReader(body))
	return string(body), true
}

type KimiAdapter struct{}

func (k KimiAdapter) ID() string { return "kimi" }

func (k KimiAdapter) BuildRequest(ctx context.Context, body map[string]any, creds *providers.Credentials) (*http.Request, error) {
	token := creds.AccessToken
	if token == "" {
		token = creds.APIKey
	}
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	return providers.NewJSONRequest(ctx, http.MethodPost, "https://api.kimi.com/coding/v1/chat/completions", body, headers)
}

func (k KimiAdapter) ParseError(resp *http.Response) (string, bool) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	resp.Body = io.NopCloser(bReader(body))
	return string(body), true
}

type DynamicOpenAIAdapter struct {
	id      string
	baseURL string
}

func (d DynamicOpenAIAdapter) ID() string { return d.id }

func (d DynamicOpenAIAdapter) BuildRequest(ctx context.Context, body map[string]any, creds *providers.Credentials) (*http.Request, error) {
	apiKey := creds.APIKey
	if apiKey == "" {
		apiKey = creds.AccessToken
	}
	headers := map[string]string{
		"Authorization": "Bearer " + apiKey,
	}
	targetURL := d.baseURL
	if !strings.Contains(targetURL, "/chat/completions") {
		if strings.HasSuffix(targetURL, "/v1") {
			targetURL += "/chat/completions"
		} else if strings.HasSuffix(targetURL, "/") {
			targetURL += "chat/completions"
		} else {
			targetURL += "/chat/completions"
		}
	}
	return providers.NewJSONRequest(ctx, http.MethodPost, targetURL, body, headers)
}

func (d DynamicOpenAIAdapter) ParseError(resp *http.Response) (string, bool) {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	resp.Body = io.NopCloser(bReader(body))
	return string(body), true
}

var KnownProviders = map[string]providers.ProviderAdapter{
	"openai":           OpenAI{},
	"anthropic":        Anthropic{},
	"claude":           Anthropic{},
	"gemini":           GeminiAdapter{},
	"github":           GitHubCopilotAdapter{},
	"kimi":             KimiAdapter{},
	"cloudflare-ai":    CloudflareAIAdapter{},
	"cf":               CloudflareAIAdapter{},
	"cerebras":         NewGenericAPIKey("cerebras", "https://api.cerebras.ai/v1", "Bearer"),
	"chutes":           NewGenericAPIKey("chutes", "https://llm.chutes.ai/v1", "Bearer"),
	"novita":           NewGenericAPIKey("novita", "https://api.novita.ai/v3/openai", "Bearer"),
	"lepton":           NewGenericAPIKey("lepton", "https://api.lepton.ai/v1", "Bearer"),
	"deepinfra":        NewGenericAPIKey("deepinfra", "https://api.deepinfra.com/v1/openai", "Bearer"),
	"minimax":          NewGenericAPIKey("minimax", "https://api.minimax.chat/v1", "Bearer"),
	"minimax-cn":       NewGenericAPIKey("minimax-cn", "https://api.minimaxi.com/v1", "Bearer"),
	"scaleway":         NewGenericAPIKey("scaleway", "https://api.scaleway.ai/v1", "Bearer"),
	"baichuan":         NewGenericAPIKey("baichuan", "https://api.baichuan-ai.com/v1", "Bearer"),
	"stepfun":          NewGenericAPIKey("stepfun", "https://api.stepfun.com/v1", "Bearer"),
	"lingyi":           NewGenericAPIKey("lingyi", "https://api.lingyiwanwu.com/v1", "Bearer"),
	"groq":             NewGenericAPIKey("groq", "https://api.groq.com/openai", "Bearer"),
	"openrouter":       NewGenericAPIKey("openrouter", "https://openrouter.ai/api", "Bearer"),
	"deepseek":         NewGenericAPIKey("deepseek", "https://api.deepseek.com", "Bearer"),
	"mistral":          NewGenericAPIKey("mistral", "https://api.mistral.ai", "Bearer"),
	"cohere":           NewGenericAPIKey("cohere", "https://api.cohere.com", "Bearer"),
	"perplexity":       NewGenericAPIKey("perplexity", "https://api.perplexity.ai", "Bearer"),
	"together":         NewGenericAPIKey("together", "https://api.together.xyz", "Bearer"),
	"fireworks":        NewGenericAPIKey("fireworks", "https://api.fireworks.ai/inference", "Bearer"),
	"huggingface":      NewGenericAPIKey("huggingface", "https://api-inference.huggingface.co", "Bearer"),
	"nvidia":           NewGenericAPIKey("nvidia", "https://integrate.api.nvidia.com", "Bearer"),
	"siliconflow":      NewGenericAPIKey("siliconflow", "https://api.siliconflow.cn", "Bearer"),
	"featherless":      NewGenericAPIKey("featherless", "https://api.featherless.ai", "Bearer"),
	"nebius":           NewGenericAPIKey("nebius", "https://api.studio.nebius.ai", "Bearer"),
	"sambanova":        NewGenericAPIKey("sambanova", "https://api.sambanova.ai", "Bearer"),
	"xai":              NewGenericAPIKey("xai", "https://api.x.ai", "Bearer"),
	"venice":           NewGenericAPIKey("venice", "https://api.venice.ai", "Bearer"),
	"hyperbolic":       NewGenericAPIKey("hyperbolic", "https://api.hyperbolic.xyz", "Bearer"),
	"kilocode":         NewGenericAPIKey("kilocode", "https://api.kilocode.com", "Bearer"),
	"cline":            NewGenericAPIKey("cline", "https://api.cline.bot", "Bearer"),
	"clinepass":        NewGenericAPIKey("clinepass", "https://api.clinepass.com", "Bearer"),
	"codebuddy-intl":   NewGenericAPIKey("codebuddy-intl", "https://api.codebuddy.ca", "Bearer"),
	"tokenrouter":      NewGenericAPIKey("tokenrouter", "https://api.tokenrouter.com", "Bearer"),
	"llm7":             NewGenericAPIKey("llm7", "https://api.llm7.io", "Bearer"),
	"morph":            NewGenericAPIKey("morph", "https://api.morph.so", "Bearer"),
	"xiaomi-tokenplan": NewGenericAPIKey("xiaomi-tokenplan", "https://api.xiaomi.com", "Bearer"),
	"bynara":           NewGenericAPIKey("bynara", "https://router.bynara.id/v1", "Bearer"),
	"geraikita":        NewGenericAPIKey("geraikita", "https://ai.geraikita.com/v1/claude", "Bearer"),
	"agnes":            NewGenericAPIKey("agnes", "https://apihub.agnes-ai.com/v1", "Bearer"),
	"bazaarlink":       NewGenericAPIKey("bazaarlink", "https://bazaarlink.ai/api/v1", "Bearer"),
	"poolside":         NewGenericAPIKey("poolside", "https://inference.poolside.ai/v1", "Bearer"),
	"api-airforce":     NewGenericAPIKey("api-airforce", "https://api.airforce/v1", "Bearer"),
	"kilo-gateway":     NewGenericAPIKey("kilo-gateway", "https://api.kilo.ai/api/gateway", "Bearer"),
}

func GetAdapter(providerID string) (providers.ProviderAdapter, bool) {
	a, ok := KnownProviders[providerID]
	return a, ok
}

func GetAdapterWithCreds(providerID string, creds *providers.Credentials) (providers.ProviderAdapter, bool) {
	if a, ok := KnownProviders[providerID]; ok {
		return a, true
	}
	if providerID == "cloudflare-ai" || providerID == "cf" {
		return CloudflareAIAdapter{}, true
	}
	if creds != nil {
		bURL := creds.BaseURL
		if bURL == "" && creds.ProviderSpecificData != nil {
			if u, ok := creds.ProviderSpecificData["baseUrl"].(string); ok && u != "" {
				bURL = u
			} else if u, ok := creds.ProviderSpecificData["baseURL"].(string); ok && u != "" {
				bURL = u
			}
		}
		if bURL != "" {
			return DynamicOpenAIAdapter{id: providerID, baseURL: bURL}, true
		}
	}
	return nil, false
}
