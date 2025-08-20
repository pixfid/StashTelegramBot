package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"path"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// BotHandler структура для обработчиков бота
type BotHandler struct {
	stash       *StashClient
	config      Config
	fileManager *FileManager
	logger      *Logger
}

func NewBotHandler(config Config) *BotHandler {
	stashClient := NewStashClient(config.StashURL, config.StashAPIKey)
	return &BotHandler{
		stash:       stashClient,
		config:      config,
		fileManager: NewFileManager(config.DATA),
		logger:      NewLogger("BotHandler"),
	}
}

// HandleStart обработчик команды /start
func (h *BotHandler) HandleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendHelp(ctx, b, update.Message.Chat.ID)
}

// HandleInfo обработчик команды /info
func (h *BotHandler) HandleInfo(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.sendHelp(ctx, b, update.Message.Chat.ID)
}

// HandleRandom обработчик команды /random
func (h *BotHandler) HandleRandom(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.logger.Info("Обработка команды /random от пользователя %d", update.Message.From.ID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "🎲 Выбираю случайное видео...",
	})

	scene, err := h.stash.GetRandomScene()
	if err != nil {
		h.logger.Error("Ошибка получения случайной сцены: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("❌ Ошибка: %v", err),
		})
		return
	}

	h.sendScene(ctx, b, update.Message.Chat.ID, scene)
}

// HandleMessage обработчик обычных сообщений
func (h *BotHandler) HandleMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	if strings.HasPrefix(update.Message.Text, "/") {
		return
	}

	h.sendHelp(ctx, b, update.Message.Chat.ID)
}

// HandleCallback обработчик callback запросов
func (h *BotHandler) HandleCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	callback := update.CallbackQuery

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
	})

	switch {
	case callback.Data == "random":
		scene, err := h.stash.GetRandomScene()
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: callback.Message.Message.Chat.ID,
				Text:   fmt.Sprintf("❌ Ошибка: %v", err),
			})
			return
		}
		h.sendScene(ctx, b, callback.Message.Message.Chat.ID, scene)

	case strings.HasPrefix(callback.Data, "performer_"):
		h.handlePerformerCallback(ctx, b, callback)
	}
}

// handlePerformerCallback обработка callback для исполнителя
func (h *BotHandler) handlePerformerCallback(ctx context.Context, b *bot.Bot, callback *models.CallbackQuery) {
	performerID := strings.TrimPrefix(callback.Data, "performer_")
	h.logger.Info("Поиск видео исполнителя: %s", performerID)

	query := `
		query FindScenes($performerID: [ID!]) {
			findScenes(
				scene_filter: {
					performers: {
						value: $performerID
						modifier: INCLUDES
					}
				}
			) {
				scenes {
					id
					title
					paths {
						screenshot
						stream
						preview
						sprite
					}
					performers {
						id
						name
					}
				}
				count
			}
		}`

	variables := map[string]interface{}{
		"performerID": []string{performerID},
	}

	resp, err := h.stash.graphQLRequest(query, variables)
	if err != nil {
		h.logger.Error("Ошибка поиска: %v", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: callback.Message.Message.Chat.ID,
			Text:   fmt.Sprintf("❌ Ошибка поиска: %v", err),
		})
		return
	}

	if len(resp.Data.FindScenes.Scenes) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: callback.Message.Message.Chat.ID,
			Text:   "❌ Видео с этим исполнителем не найдены",
		})
		return
	}

	randomIndex := rand.Intn(len(resp.Data.FindScenes.Scenes))
	scene := &resp.Data.FindScenes.Scenes[randomIndex]

	h.sendScene(ctx, b, callback.Message.Message.Chat.ID, scene)
}

// sendScene отправляет сцену
func (h *BotHandler) sendScene(ctx context.Context, b *bot.Bot, chatID int64, scene *Scene) {
	h.logger.Info("Отправка сцены: %s", scene.Title)

	previewURL := fmt.Sprintf("%s?apikey=%s", scene.Paths.Preview, h.config.StashAPIKey)

	filepath, err := h.fileManager.DownloadFile(previewURL, scene.Title)
	if err != nil {
		h.logger.Error("Не удалось загрузить превью: %v", err)
		h.sendSceneWithoutPreview(ctx, b, chatID, scene)
		return
	}

	fileData, err := h.fileManager.ReadFile(filepath)
	if err != nil {
		h.logger.Error("Не удалось прочитать файл: %v", err)
		h.sendSceneWithoutPreview(ctx, b, chatID, scene)
		return
	}

	streamURL := fmt.Sprintf("%s", scene.Paths.Stream)
	kb := CreateSceneKeyboard(scene, streamURL)
	caption := fmt.Sprintf("🎬 <b>%s</b>", escapeHTML(scene.Title))

	params := &bot.SendDocumentParams{
		ChatID: chatID,
		Document: &models.InputFileUpload{
			Filename: path.Base(filepath),
			Data:     bytes.NewReader(fileData),
		},
		Caption:     caption,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	}

	_, err = b.SendDocument(ctx, params)
	if err != nil {
		h.logger.Error("Не удалось отправить документ: %v", err)
		h.sendSceneWithoutPreview(ctx, b, chatID, scene)
	} else {
		h.logger.Success("Сцена отправлена успешно")
	}

	h.fileManager.DeleteFile(filepath)
}

// sendSceneWithoutPreview отправляет сцену без превью
func (h *BotHandler) sendSceneWithoutPreview(ctx context.Context, b *bot.Bot, chatID int64, scene *Scene) {
	h.logger.Warning("Отправка без превью")

	streamURL := fmt.Sprintf("%s", scene.Paths.Stream)
	kb := CreateSceneKeyboard(scene, streamURL)
	text := fmt.Sprintf("🎬 <b>%s</b>", escapeHTML(scene.Title))

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	})
}

// sendHelp отправляет справку
func (h *BotHandler) sendHelp(ctx context.Context, b *bot.Bot, chatID int64) {
	helpText := `🎬 <b>StashApp Bot</b>

<b>Доступные команды:</b>

🎲 /random - Случайное видео
ℹ️ /info - Информация о боте
❓ /start - Начать работу

💡 <i>Используйте /random для получения случайного видео</i>`

	kb := CreateHelpKeyboard()

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        helpText,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: bot.True(),
		},
	})
}
