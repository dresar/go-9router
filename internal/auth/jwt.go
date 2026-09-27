package auth

import (
	"errors"
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
		return "", errors.New("JWT_SECRET is not set")
	}
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(sessionDuration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifySession(tokenStr, secret string) error {
	if secret == "" {
		return errors.New("JWT_SECRET is not set")
	}
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid token")
	}
	return nil
}

func CookieName() string {
	return cookieName
}
