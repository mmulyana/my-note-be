package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL    string
	Port           string
	AllowedOrigins []string
	JWTSecret      string
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

	return &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		Port:           port,
		AllowedOrigins: strings.Split(origins, ","),
		JWTSecret:      jwtSecret,
	}
}
