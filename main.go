package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	// Инициализация
	rand.Seed(time.Now().UnixNano())
	logger := NewLogger("Main")

	// Загружаем конфигурацию
	config := LoadConfig()

	// Проверяем подключение к StashApp
	logger.Info("Проверка подключения к StashApp: %s", config.StashURL)
	testClient := NewStashClient(config.StashURL, config.StashAPIKey)

	if err := testClient.TestConnection(); err != nil {
		logger.Warning("Не удалось подключиться к StashApp: %v", err)
		logger.Info("Бот запущен, но может не работать до устранения проблемы")
	} else {
		logger.Success("Успешно подключено к StashApp!")
	}

	// Создаем обработчик
	handler := NewBotHandler(config)

	// Создаем бота
	opts := []bot.Option{
		bot.WithDefaultHandler(handler.HandleMessage),
		bot.WithCallbackQueryDataHandler("", bot.MatchTypePrefix, handler.HandleCallback),
	}

	b, err := bot.New(config.TelegramToken, opts...)
	if err != nil {
		logger.Error("Не удалось создать бота: %v", err)
		panic(err)
	}

	// Регистрируем команды
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, handler.HandleStart)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/info", bot.MatchTypeExact, handler.HandleInfo)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/random", bot.MatchTypeExact, handler.HandleRandom)

	// Устанавливаем команды в меню бота
	b.SetMyCommands(context.Background(), &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "start", Description: "Начать работу с ботом"},
			{Command: "info", Description: "Информация о боте"},
			{Command: "random", Description: "Случайное видео"},
		},
	})

	// Создаем контекст для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Получаем информацию о боте
	me, err := b.GetMe(context.Background())

	if err != nil {
		logger.Warning("Не удалось получить информацию о боте: %v", err)
	} else {
		logger.Success("Bot username: @%s", me.Username)
	}

	// Запускаем бота
	logger.Success("Бот запущен успешно!")
	b.Start(ctx)
}
