package bot

import "github.com/go-telegram/bot/models"

func BuildReplyKeyboard() models.ReplyKeyboardMarkup {
	subBtn := models.KeyboardButton{Text: "✅ Подписаться"}
	unsubBtn := models.KeyboardButton{Text: "❌ Отписаться"}
	statusBtn := models.KeyboardButton{Text: "❔ Статус"}
	helpBtn := models.KeyboardButton{Text: "📋 Помощь"}

	return models.ReplyKeyboardMarkup{
		Keyboard:        [][]models.KeyboardButton{{subBtn, unsubBtn}, {statusBtn, helpBtn}},
		IsPersistent:    true,
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
		Selective:       false,
	}
}
