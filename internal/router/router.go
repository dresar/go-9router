package router

import (
	"net/http"

	"github.com/dresar/go-9router/internal/handlers"
	"github.com/dresar/go-9router/internal/middleware"
)

func New(h *handlers.Handler, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", h.HandleHealth)
	mux.HandleFunc("/api/version", h.HandleVersion)
	mux.HandleFunc("/api/init", h.HandleInit)
	mux.HandleFunc("/api/locale", h.HandleLocale)
	mux.HandleFunc("/api/shutdown", h.HandleShutdown)

	mux.HandleFunc("/api/auth/login", h.HandleAuthLogin)
	mux.HandleFunc("/api/auth/logout", h.HandleAuthLogout)
	mux.HandleFunc("/api/auth/status", h.HandleAuthStatus)

	mux.HandleFunc("/api/auth/oidc/", notImplemented)
	mux.HandleFunc("/api/auth/saml/", notImplemented)

	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			middleware.RequireSession(jwtSecret, http.HandlerFunc(next)).ServeHTTP(w, r)
		}
	}

	mux.HandleFunc("/api/settings", auth(h.HandleSettings))

	mux.HandleFunc("/api/providers", auth(h.HandleProviders))
	mux.HandleFunc("/api/providers/", auth(h.HandleProviderByID))

	mux.HandleFunc("/api/provider-nodes", auth(h.HandleNodes))
	mux.HandleFunc("/api/provider-nodes/", auth(h.HandleNodeByID))

	mux.HandleFunc("/api/keys", auth(h.HandleKeys))
	mux.HandleFunc("/api/keys/", auth(h.HandleKeyByID))

	mux.HandleFunc("/api/combos", auth(h.HandleCombos))
	mux.HandleFunc("/api/combos/", auth(h.HandleComboByID))

	mux.HandleFunc("/api/proxy-pools", auth(h.HandleProxyPools))
	mux.HandleFunc("/api/proxy-pools/", auth(h.HandleProxyPoolByID))

	mux.HandleFunc("/api/models", auth(h.HandleModels))
	mux.HandleFunc("/api/models/alias", auth(h.HandleModelAliases))
	mux.HandleFunc("/api/models/custom", auth(h.HandleCustomModels))

	mux.HandleFunc("/api/usage", auth(h.HandleUsageStats))
	mux.HandleFunc("/api/usage/history", auth(h.HandleUsageHistory))
	mux.HandleFunc("/api/usage/stream", h.HandleUsageStream)
	mux.HandleFunc("/api/usage/request-details", auth(h.HandleRequestDetails))
	mux.HandleFunc("/api/usage/stats", auth(h.HandleUsageStats))

	mux.HandleFunc("/api/oauth/", notImplemented)
	mux.HandleFunc("/api/mcp/", notImplemented)
	mux.HandleFunc("/api/headroom/", h.HandleHeadroomStatus)
	mux.HandleFunc("/api/pxpipe/", notImplemented)
	mux.HandleFunc("/api/tunnel/", h.HandleTunnelStatus)
	mux.HandleFunc("/api/translator/", h.HandleTranslator)
	mux.HandleFunc("/api/cli-tools/", h.HandleCLITools)
	mux.HandleFunc("/api/media-providers/", notImplemented)
	mux.HandleFunc("/api/tags", h.HandleTags)
	mux.HandleFunc("/api/pricing", h.HandlePricing)

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
