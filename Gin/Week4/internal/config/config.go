package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RedisURL           string
	KafkaBroker        string
	KafkaTopic         string
	APIPort            string
	CollectorPort      string
	RateLimitRequests  int
	RateLimitWindow    int
	NewsAPIKey         string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load("config.env")

	rateLimitRequests, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	rateLimitWindow, _ := strconv.Atoi(getEnv("RATE_LIMIT_WINDOW", "60"))

	return &Config{
		RedisURL:          getEnv("REDIS_URL", "localhost:6379"),
		KafkaBroker:       getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:        getEnv("KAFKA_TOPIC", "news-updates"),
		APIPort:           getEnv("API_PORT", "8080"),
		CollectorPort:     getEnv("COLLECTOR_PORT", "8081"),
		RateLimitRequests: rateLimitRequests,
		RateLimitWindow:   rateLimitWindow,
		NewsAPIKey:        getEnv("NEWS_API_KEY", ""),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
