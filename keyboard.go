package main

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

// CreateSceneKeyboard создает клавиатуру для сцены
func CreateSceneKeyboard(scene *Scene, streamURL string) *models.InlineKeyboardMarkup {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{},
	}

	// Кнопки исполнителей
	if len(scene.Performers) > 0 {
		performerRow := []models.InlineKeyboardButton{}

		for i, performer := range scene.Performers {
			// Максимум 2 кнопки в ряд
			if i > 0 && i%2 == 0 {
				kb.InlineKeyboard = append(kb.InlineKeyboard, performerRow)
				performerRow = []models.InlineKeyboardButton{}
			}

			performerRow = append(performerRow, models.InlineKeyboardButton{
				Text:         fmt.Sprintf("🔍 %s", performer.Name),
				CallbackData: fmt.Sprintf("performer_%s", performer.ID),
			})
		}

		if len(performerRow) > 0 {
			kb.InlineKeyboard = append(kb.InlineKeyboard, performerRow)
		}
	}
	// Кнопка студии
	if scene.Studio.Name != "" {
		studioRow := []models.InlineKeyboardButton{}
		studioRow = append(studioRow, models.InlineKeyboardButton{
			Text:         fmt.Sprintf("📹 %s", scene.Studio.Name),
			CallbackData: fmt.Sprintf("studio_%s", scene.Studio.ID),
		})
		if studioRow != nil {
			kb.InlineKeyboard = append(kb.InlineKeyboard, studioRow)
		}
	}

	// Кнопка стрима
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{
			Text: "🔗 Открыть стрим",
			URL:  streamURL,
		},
	})

	// Кнопка случайного видео
	kb.InlineKeyboard = append(kb.InlineKeyboard, []models.InlineKeyboardButton{
		{
			Text:         "🎲 Случайное",
			CallbackData: "random",
		},
	})

	return kb
}

// CreateHelpKeyboard создает клавиатуру для справки
func CreateHelpKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🎲 Случайное видео", CallbackData: "random"},
			},
		},
	}
}
