package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port    string
	GinMode string

	// Database
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string

	// Redis
	RedisURL      string
	RedisPassword string

	// JWT
	JWTPrivateKey    string
	JWTPublicKey     string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	// CORS
	AllowedOrigins []string

	// S3/MinIO
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool

	// TURN Server
	TURNServerURL string
	TURNUsername  string
	TURNPassword  string

	// Rate Limiting
	RateLimitRequests int
	RateLimitDuration time.Duration

	// File Upload
	MaxFileSize int64

	// Cookie
	CookieSecure   bool
	CookieSameSite string
}

func Load() (*Config, error) {
	// Load .env file
	godotenv.Load()

	cfg := &Config{
		Port:    getEnv("PORT", "8080"),
		GinMode: getEnv("GIN_MODE", "debug"),

		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", "postgres"),
		DBName:      getEnv("DB_NAME", "slack_clone"),

		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		JWTPrivateKey:    getEnv("JWT_PRIVATE_KEY", ""),
		JWTPublicKey:     getEnv("JWT_PUBLIC_KEY", ""),
		JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),

		S3Endpoint:  getEnv("S3_ENDPOINT", "localhost:9000"),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "minioadmin"),
		S3SecretKey: getEnv("S3_SECRET_KEY", "minioadmin"),
		S3Bucket:    getEnv("S3_BUCKET", "slack-clone"),
		S3UseSSL:    getEnv("S3_USE_SSL", "false") == "true",

		TURNServerURL: getEnv("TURN_SERVER_URL", "turn:localhost:3478"),
		TURNUsername:  getEnv("TURN_USERNAME", "turnuser"),
		TURNPassword:  getEnv("TURN_PASSWORD", "turnpass"),

		MaxFileSize: 52428800, // 50MB

		CookieSecure:   getEnv("COOKIE_SECURE", "false") == "true",
		CookieSameSite: getEnv("COOKIE_SAME_SITE", "lax"),
	}

	// Parse allowed origins
	origins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	cfg.AllowedOrigins = parseCommaSeparated(origins)

	// Build DATABASE_URL if not provided
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
		)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}

	var result []string
	for _, item := range split(s, ',') {
		trimmed := trim(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func split(s string, sep rune) []string {
	var result []string
	current := ""

	for _, char := range s {
		if char == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)

	return result
}

func trim(s string) string {
	start := 0
	end := len(s) - 1

	for start <= end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}

	for end >= start && (s[end] == ' ' || s[end] == '\t' || s[end] == '\n') {
		end--
	}

	if start > end {
		return ""
	}

	return s[start : end+1]
}
