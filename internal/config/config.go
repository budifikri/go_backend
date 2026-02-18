package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Email    EmailConfig
	CORS     CORSConfig
	Log      LogConfig
}

type ServerConfig struct {
	Host string
	Port int
	Env  string
}

type DatabaseConfig struct {
	Host           string
	Port           int
	Name           string
	User           string
	Password       string
	MaxConnections int
	SSLMode        string
}

type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

type EmailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// .env file is optional in production
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("APP_HOST", "0.0.0.0"),
			Port: getEnvAsInt("APP_PORT", 3000),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnvAsInt("DB_PORT", 5432),
			Name:           getEnv("DB_NAME", "pos_retail"),
			User:           getEnv("DB_USER", "postgres"),
			Password:       getEnv("DB_PASSWORD", "postgres"),
			MaxConnections: getEnvAsInt("DB_MAX_CONNECTIONS", 10),
			SSLMode:        getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpiresIn: getEnvAsDuration("JWT_EXPIRES_IN", 8*time.Hour),
		},
		Email: EmailConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@pos-retail.com"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", ","),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", ","),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", ","),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key, separator string) []string {
	value := getEnv(key, "")
	if value == "" {
		return []string{}
	}
	return splitAndTrim(value, separator)
}

func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range split(s, sep) {
		trimmed := trim(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func split(s, sep string) []string {
	result := make([]string, 0)
	current := ""
	for _, ch := range s {
		if string(ch) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	result = append(result, current)
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
