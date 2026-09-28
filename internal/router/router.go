package router

import (
	"net/http"
	"strings"

	"github.com/dresar/go-9router/internal/handlers"
	"github.com/dresar/go-9router/internal/middleware"
)

func New(h *handlers.Handler, jwtSecret string) http.Handler {
	if jwtSecret == "" {
		jwtSecret = "9router-default-jwt-session-secret-key-2026"
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", h.HandleHealth)
	mux.HandleFunc("/api/version", h.HandleVersion)
	mux.HandleFunc("/api/version/check", h.HandleVersionCheck)
	mux.HandleFunc("/api/version/sync", h.HandleVersionSync)
	mux.HandleFunc("/api/version/shutdown", h.HandleShutdown)
	mux.HandleFunc("/api/version/update", h.HandleVersionActions)
	mux.HandleFunc("/api/init", h.HandleInit)
	mux.HandleFunc("/api/locale", h.HandleLocale)
	mux.HandleFunc("/api/shutdown", h.HandleShutdown)

	mux.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("/api/auth/login", h.HandleAuthLogin)
	mux.HandleFunc("/api/auth/logout", h.HandleAuthLogout)
	mux.HandleFunc("/api/auth/status", h.HandleAuthStatus)
	mux.HandleFunc("/api/auth/reset-password", h.HandleAuthResetPassword)

	mux.HandleFunc("/api/auth/oidc", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"configured": false})
	})
	mux.HandleFunc("/api/auth/oidc/", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"configured": false})
	})
	mux.HandleFunc("/api/auth/saml", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"configured": false})
	})
	mux.HandleFunc("/api/auth/saml/", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"configured": false})
	})

	mux.HandleFunc("/api/cloud", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"connected": false})
	})

	mux.HandleFunc("/api/dashboard/chat/completions", h.HandleChat)

	mux.HandleFunc("/api/settings/require-login", h.HandleRequireLogin)

	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			middleware.RequireSession(jwtSecret, h.DB, http.HandlerFunc(next)).ServeHTTP(w, r)
		}
	}

	mux.HandleFunc("/api/settings", auth(h.HandleSettings))
	mux.HandleFunc("/api/settings/database", auth(h.HandleSettingsDatabase))
	mux.HandleFunc("/api/settings/proxy-test", auth(h.HandleSettingsProxyTest))

	mux.HandleFunc("/api/providers", auth(h.HandleProviders))
	mux.HandleFunc("/api/providers/client", auth(h.HandleProvidersClient))
	mux.HandleFunc("/api/providers/suggested-models", auth(h.HandleSuggestedModels))
	mux.HandleFunc("/api/providers/test-batch", auth(h.HandleTestBatch))
	mux.HandleFunc("/api/providers/validate", auth(func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"valid": true})
	}))
	mux.HandleFunc("/api/providers/", auth(h.HandleProviderByID))

	mux.HandleFunc("/api/provider-nodes", auth(h.HandleNodes))
	mux.HandleFunc("/api/provider-nodes/validate", auth(h.HandleNodeValidate))
	mux.HandleFunc("/api/provider-nodes/", auth(h.HandleNodeByID))

	mux.HandleFunc("/api/keys", auth(h.HandleKeys))
	mux.HandleFunc("/api/keys/", auth(h.HandleKeyByID))

	mux.HandleFunc("/api/combos", auth(h.HandleCombos))
	mux.HandleFunc("/api/combos/presets", auth(h.HandleComboPresets))
	mux.HandleFunc("/api/combos/", auth(h.HandleComboByID))

	mux.HandleFunc("/api/proxy-pools", auth(h.HandleProxyPools))
	mux.HandleFunc("/api/proxy-pools/cloudflare-deploy", auth(h.HandleProxyPoolDeploy))
	mux.HandleFunc("/api/proxy-pools/deno-deploy", auth(h.HandleProxyPoolDeploy))
	mux.HandleFunc("/api/proxy-pools/vercel-deploy", auth(h.HandleProxyPoolDeploy))
	mux.HandleFunc("/api/proxy-pools/", auth(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/test") {
			h.HandleProxyPoolTest(w, r)
			return
		}
		h.HandleProxyPoolByID(w, r)
	}))

	mux.HandleFunc("/api/models", auth(h.HandleModels))
	mux.HandleFunc("/api/models/alias", auth(h.HandleModelAliases))
	mux.HandleFunc("/api/models/custom", auth(h.HandleCustomModels))
	mux.HandleFunc("/api/models/availability", auth(h.HandleModelAvailability))
	mux.HandleFunc("/api/models/disabled", auth(h.HandleDisabledModels))
	mux.HandleFunc("/api/models/test", auth(h.HandleModelTest))

	mux.HandleFunc("/api/usage", auth(h.HandleUsageStats))
	mux.HandleFunc("/api/usage/history", auth(h.HandleUsageHistory))
	mux.HandleFunc("/api/usage/stream", h.HandleUsageStream)
	mux.HandleFunc("/api/usage/request-details", auth(h.HandleRequestDetails))
	mux.HandleFunc("/api/usage/request-logs", auth(h.HandleRequestLogs))
	mux.HandleFunc("/api/usage/chart", auth(h.HandleUsageChart))
	mux.HandleFunc("/api/usage/providers", auth(h.HandleUsageProviders))
	mux.HandleFunc("/api/usage/logs", auth(h.HandleUsageLogs))
	mux.HandleFunc("/api/usage/stats", auth(h.HandleUsageStats))
	mux.HandleFunc("/api/usage/", auth(h.HandleUsageConnectionSub))

	mux.HandleFunc("/api/oauth", h.HandleOAuth)
	mux.HandleFunc("/api/oauth/", h.HandleOAuth)
	mux.HandleFunc("/callback", h.HandleOAuthCallback)
	mux.HandleFunc("/callback/", h.HandleOAuthCallback)
	mux.HandleFunc("/api/mcp", h.HandleMCPProxy)
	mux.HandleFunc("/api/mcp/", h.HandleMCPProxy)
	mux.HandleFunc("/api/token-saver/test", h.HandleTokenSaverTest)
	mux.HandleFunc("/api/token-saver/summary", h.HandleTokenSaverSummary)
	mux.HandleFunc("/api/headroom", h.HandleHeadroomStatus)
	mux.HandleFunc("/api/headroom/", h.HandleHeadroomStatus)
	mux.HandleFunc("/api/pxpipe", h.HandlePxpipe)
	mux.HandleFunc("/api/pxpipe/", h.HandlePxpipe)
	mux.HandleFunc("/api/tunnel", h.HandleTunnelStatus)
	mux.HandleFunc("/api/tunnel/status", h.HandleTunnelStatus)
	mux.HandleFunc("/api/tunnel/", h.HandleTunnelAction)
	mux.HandleFunc("/api/translator", h.HandleTranslator)
	mux.HandleFunc("/api/translator/", h.HandleTranslator)
	mux.HandleFunc("/api/cli-tools", h.HandleCLITools)
	mux.HandleFunc("/api/cli-tools/", h.HandleCLITools)
	mux.HandleFunc("/api/media-providers", h.HandleMediaProviders)
	mux.HandleFunc("/api/media-providers/", h.HandleMediaProviders)
	mux.HandleFunc("/api/tags", h.HandleTags)
	mux.HandleFunc("/api/pricing", h.HandlePricing)

	mux.HandleFunc("/api/changelog", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, "CHANGELOG.md")
	})

	mux.HandleFunc("/api/payments", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"payments": []any{}})
	})
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"users": []any{}})
	})

	// /v1 routes
	mux.HandleFunc("/v1/chat/completions", h.HandleChat)
	mux.HandleFunc("/v1/messages", h.HandleMessages)
	mux.HandleFunc("/v1/models", h.HandleV1Models)
	mux.HandleFunc("/v1/embeddings", h.HandleEmbeddings)
	mux.HandleFunc("/v1/images/generations", h.HandleImagesGenerations)
	mux.HandleFunc("/v1/audio/speech", h.HandleAudioSpeech)
	mux.HandleFunc("/v1/audio/transcriptions", h.HandleAudioTranscriptions)
	mux.HandleFunc("/v1/search", h.HandleSearch)
	mux.HandleFunc("/v1/web/fetch", h.HandleWebFetch)
	mux.HandleFunc("/v1/videos/generations", h.HandleVideoGenerations)
	mux.HandleFunc("/v1/videos/", h.HandleVideoStatus)
	mux.HandleFunc("/v1beta/", notImplemented)

	// /api/v1 routes
	mux.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "go-9router"})
	})
	mux.HandleFunc("/api/v1/chat/completions", h.HandleChat)
	mux.HandleFunc("/api/v1/messages", h.HandleMessages)
	mux.HandleFunc("/api/v1/models", h.HandleV1Models)
	mux.HandleFunc("/api/v1/embeddings", h.HandleEmbeddings)
	mux.HandleFunc("/api/v1/images/generations", h.HandleImagesGenerations)
	mux.HandleFunc("/api/v1/audio/speech", h.HandleAudioSpeech)
	mux.HandleFunc("/api/v1/audio/transcriptions", h.HandleAudioTranscriptions)
	mux.HandleFunc("/api/v1/search", h.HandleSearch)
	mux.HandleFunc("/api/v1/web/fetch", h.HandleWebFetch)
	mux.HandleFunc("/api/v1beta", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "go-9router-beta"})
	})
	mux.HandleFunc("/api/v1beta/", func(w http.ResponseWriter, r *http.Request) {
		h.JSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "go-9router-beta"})
	})

	mux.HandleFunc("/codex/", h.HandleChat)
	mux.HandleFunc("/responses", h.HandleChat)

	mux.HandleFunc("/", h.HandleFrontend)

	return middleware.CORS(middleware.Logging(mux))
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"this endpoint is not yet implemented in go-9router"}`))
}
