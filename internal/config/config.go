package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                string
	DataDir             string
	JWTSecret           string
	InitialPassword     string
	APIKeySecret        string
	MachineIDSalt       string
	EnableRequestLogs   bool
	ObservabilityEnabled bool
	AuthCookieSecure    bool
	RequireAPIKey       bool
	BaseURL             string
	CloudURL            string
	SearxNGURL          string
	HTTPProxy           string
	HTTPSProxy          string
	AllProxy            string
	NoProxy             string

	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	StreamTimeout     time.Duration
	MaxBodyBytes      int64
	MaxConcurrent     int
}

func Load() *Config {
	return &Config{
		Port:                 envStr("PORT", "20128"),
		DataDir:              envStr("DATA_DIR", "data"),
		JWTSecret:            envStr("JWT_SECRET", "9router-default-jwt-session-secret-key-2026"),
		InitialPassword:      envStr("INITIAL_PASSWORD", "admin1234"),
		APIKeySecret:         envStr("API_KEY_SECRET", ""),
		MachineIDSalt:        envStr("MACHINE_ID_SALT", ""),
		EnableRequestLogs:    envBool("ENABLE_REQUEST_LOGS", false),
		ObservabilityEnabled: envBool("OBSERVABILITY_ENABLED", true),
		AuthCookieSecure:     envBool("AUTH_COOKIE_SECURE", false),
		RequireAPIKey:        envBool("REQUIRE_API_KEY", false),
		BaseURL:              envStr("BASE_URL", "http://localhost:20128"),
		CloudURL:             envStr("CLOUD_URL", "https://9router.com"),
		SearxNGURL:           envStr("SEARXNG_URL", ""),
		HTTPProxy:            envStr("HTTP_PROXY", envStr("http_proxy", "")),
		HTTPSProxy:           envStr("HTTPS_PROXY", envStr("https_proxy", "")),
		AllProxy:             envStr("ALL_PROXY", envStr("all_proxy", "")),
		NoProxy:              envStr("NO_PROXY", envStr("no_proxy", "")),

		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       120 * time.Second,
		StreamTimeout:     0,
		MaxBodyBytes:      int64(envInt("MAX_BODY_MB", 128)) * 1024 * 1024,
		MaxConcurrent:     envInt("MAX_CONCURRENT", 500),
	}
}

func envStr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
