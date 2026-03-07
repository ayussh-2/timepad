package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddr       string
	DatabaseURL      string
	RedisURL         string
	JWTPrivateKey    string
	JWTPublicKey     string
	JWTAccessExpiry  int
	JWTRefreshExpiry int
	Env              string
	RateLimitRPM     int
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		ServerAddr:       getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/timepad?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTPrivateKey:    getEnv("JWT_PRIVATE_KEY", getEnv("JWT_PRIVATE_KEY_FILE", "./secrets/private.pem")),
		JWTPublicKey:     getEnv("JWT_PUBLIC_KEY", getEnv("JWT_PUBLIC_KEY_FILE", "./secrets/public.pem")),
		JWTAccessExpiry:  getEnvInt("JWT_ACCESS_EXPIRY", 3600),
		JWTRefreshExpiry: getEnvInt("JWT_REFRESH_EXPIRY", 2592000),
		Env:              getEnv("ENV", "development"),
		RateLimitRPM:     getEnvInt("RATE_LIMIT_RPM", 60),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
