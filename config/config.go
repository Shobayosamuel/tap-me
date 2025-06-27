package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host string
	Port string
	User string
	Password string
	DBName string
	SSLMode string
}

type JWTConfig struct {
	Secret            string
	RefreshSecret     string
	AccessTokenTTL    int // hours
	RefreshTokenTTL   int // days
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "tapme_user"),
			Password: getEnv("DB_PASSWORD", "Shobayo78"),
			DBName:   getEnv("DB_NAME", "tapme-db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-jwt-secret"),
			RefreshSecret:   getEnv("REFRESH_SECRET", "your-refresh-secret"),
			AccessTokenTTL:  getEnvAsInt("ACCESS_TOKEN_TTL_HOURS", 1),
			RefreshTokenTTL: getEnvAsInt("REFRESH_TOKEN_TTL_DAYS", 7),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}