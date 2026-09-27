package middleware

import (
	"database/sql"
	"net"
	"net/http"
	"strings"

	"github.com/dresar/go-9router/internal/auth"
	"github.com/dresar/go-9router/internal/storage/repos"
)

func RequireSession(secret string, db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractSessionToken(r)
		if token != "" && auth.VerifySession(token, secret) == nil {
			next.ServeHTTP(w, r)
			return
		}

		if db != nil {
			settings, err := repos.GetSettings(db)
			if err == nil && !repos.SettingBool(settings, "requireLogin", false) {
				if isLocalRequest(r) {
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
	})
}

func extractSessionToken(r *http.Request) string {
	for _, cName := range []string{"9r_session", "auth_token", "session"} {
		if c, err := r.Cookie(cName); err == nil && c.Value != "" {
			return c.Value
		}
	}
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:]
	}
	return ""
}

func isLocalRequest(r *http.Request) bool {
	host := r.Host
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err == nil {
			host = h
		}
	}
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	remote := r.RemoteAddr
	if strings.Contains(remote, ":") {
		h, _, err := net.SplitHostPort(remote)
		if err == nil {
			remote = h
		}
	}
	return remote == "127.0.0.1" || remote == "::1" || remote == "localhost"
}
