package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BINGX_API_KEY    string
	BINGX_API_SECRET string
	CHAT_ID          string
	CHAT_BOT_TOKEN   string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	cfg := &Config{
		BINGX_API_KEY:    getEnv("BINGX_API_KEY"),
		BINGX_API_SECRET: getEnv("BINGX_API_SECRET"),
		CHAT_ID:          getEnv("CHAT_ID"),
		CHAT_BOT_TOKEN:   getEnv("CHAT_BOT_TOKEN"),
	}

	log.Printf("Config loaded successfully")
	// log.Printf(" DB: %s@%s:%d/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
	// log.Printf(" Server: http://localhost%s", cfg.ServerPort)

	return cfg
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("❌ Required environment variable %s is not set", key)
	}
	return value
}
