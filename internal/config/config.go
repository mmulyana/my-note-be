package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL     string
	Port            string
	AllowedOrigins  []string
	JWTSecret       string
	RelayHubBaseURL string
	RelayHubToken   string
	OpenRouterKey   string
	OpenRouterModel string
	AIDailyTokens   int
	// note: allowlist of who can author release notes; unset means nobody can
	AdminEmails []string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:5173"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key"
	}

	relayHubBaseURL := os.Getenv("RELAY_HUB_BASE_URL")
	if relayHubBaseURL == "" {
		relayHubBaseURL = "http://localhost:8080"
	}

	aiDailyTokens := 200000
	if raw := os.Getenv("AI_DAILY_TOKEN_LIMIT"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			aiDailyTokens = n
		}
	}

	var adminEmails []string
	if raw := os.Getenv("ADMIN_EMAILS"); raw != "" {
		adminEmails = strings.Split(raw, ",")
	}

	return &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		Port:            port,
		AllowedOrigins:  strings.Split(origins, ","),
		JWTSecret:       jwtSecret,
		RelayHubBaseURL: relayHubBaseURL,
		RelayHubToken:   os.Getenv("RELAY_HUB_TOKEN"),
		OpenRouterKey:   os.Getenv("OPENROUTER_API_KEY"),
		OpenRouterModel: os.Getenv("OPENROUTER_MODEL"),
		AIDailyTokens:   aiDailyTokens,
		AdminEmails:     adminEmails,
	}
}
