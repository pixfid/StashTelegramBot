package main

import (
	"log"
	"os"
	"strings"
)

// Config структура для конфигурации
type Config struct {
	TelegramToken string
	StashURL      string
	StashAPIKey   string
	DATA          string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() Config {

	config := Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		StashURL:      os.Getenv("STASH_URL"),
		StashAPIKey:   os.Getenv("STASH_API_KEY"),
		DATA:          os.Getenv("DATA"),
	}

	if config.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	if config.StashURL == "" {
		log.Fatal("STASH_URL не установлен")
	}

	if config.StashAPIKey == "" {
		log.Fatal("STASH_API_KEY не установлен")
	}

	if config.DATA == "" {
		log.Fatal("DATA не установлен")
	}

	config.StashURL = strings.TrimSuffix(config.StashURL, "/")

	return config
}
