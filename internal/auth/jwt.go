package auth

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	sessionDuration = 30 * 24 * time.Hour
	cookieName      = "9r_session"
)

type Claims struct {
	jwt.RegisteredClaims
}

func SignSession(secret string) (string, error) {
	if secret == "" {
		secret = "9router-default-jwt-session-secret-key-2026"
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"authenticated": true,
		"iat":           now.Unix(),
		"exp":           now.Add(sessionDuration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifySession(tokenStr, secret string) error {
	if strings.TrimSpace(tokenStr) == "" {
		return errors.New("empty token")
	}

	secrets := []string{}
	if secret != "" {
		secrets = append(secrets, secret)
	}
	if secret != "9router-default-jwt-session-secret-key-2026" {
		secrets = append(secrets, "9router-default-jwt-session-secret-key-2026")
	}
	for _, p := range []string{"data/jwt-secret", "../data/jwt-secret", "frontend/data/jwt-secret"} {
		if b, err := os.ReadFile(p); err == nil {
			s := strings.TrimSpace(string(b))
			if s != "" {
				secrets = append(secrets, s)
			}
		}
	}

	for _, sec := range secrets {
		if sec == "" {
			continue
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(sec), nil
		})
		if err == nil && token.Valid {
			return nil
		}
	}

	return errors.New("invalid token")
}

func CookieName() string {
	return cookieName
}
