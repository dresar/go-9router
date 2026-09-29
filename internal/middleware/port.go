package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type PortContextKey string

const (
	ServerPortKey     PortContextKey = "server_port"
	IsDirectNoAuthKey PortContextKey = "is_direct_no_auth"
)

// WithServerPort wraps a handler and attaches the port and direct no-auth status to the request context.
func WithServerPort(port int, isDirect bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ServerPortKey, port)
			ctx = context.WithValue(ctx, IsDirectNoAuthKey, isDirect)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// IsDirectNoAuthRequest checks if the incoming request is allowed without an API key.
func IsDirectNoAuthRequest(r *http.Request, noAuthPorts []string) bool {
	// 1. Direct server flag in context
	if isDirect, ok := r.Context().Value(IsDirectNoAuthKey).(bool); ok && isDirect {
		return true
	}

	// 2. Server port from context
	if port, ok := r.Context().Value(ServerPortKey).(int); ok {
		pStr := strconv.Itoa(port)
		for _, nap := range noAuthPorts {
			if strings.TrimSpace(nap) == pStr {
				return true
			}
		}
	}

	// 3. Port from Host header
	host := r.Host
	if strings.Contains(host, ":") {
		_, p, err := net.SplitHostPort(host)
		if err == nil && p != "" {
			for _, nap := range noAuthPorts {
				if strings.TrimSpace(nap) == p {
					return true
				}
			}
		}
	}

	return false
}
