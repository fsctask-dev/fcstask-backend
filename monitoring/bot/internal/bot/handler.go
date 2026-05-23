package bot

import (
	"context"
	"time"

	"fcstask-monitor-bot/internal/db"
	"fcstask-monitor-bot/internal/logger"
	model "fcstask-monitor-bot/internal/model"

	gotgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"gorm.io/gorm"
)

func Start(ctx context.Context, bot *gotgbot.Bot, update *models.Update) {
	if update.Message == nil {
		logger.Log.Debug().Int64("update_id", update.ID).Msg("update message is nil")
		return
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text

	logger.Log.Info().Int64("chat_id", chatID).Str("text", text).Msg("start command received")

	user := model.User{
		ChatID:    chatID,
		CreatedAt: time.Now(),
	}

	res := database.DB.Where("chat_id = ?", chatID).FirstOrCreate(&user)
	switch {
	case res.Error != nil:
		logger.Log.Error().Err(res.Error).Int64("chat_id", chatID).Msg("failed to create user")
	case res.RowsAffected == 0:
		_, err := bot.SendMessage(ctx, &gotgbot.SendMessageParams{
			ChatID: chatID,
			Text:   "⏹️ Вы уже подписаны на алёрты",
		})
		logger.Log.Debug().Err(err).Int64("chat_id", chatID).Msg("sent already subscribed message")
	default:
		logger.Log.Info().Int64("chat_id", chatID).Msg("new subscription created")
		_, err := bot.SendMessage(ctx, &gotgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "✅Вы подписались на алёрты",
			ReplyMarkup: BuildReplyKeyboard(),
		})
		logger.Log.Debug().Err(err).Int64("chat_id", chatID).Msg("sent subscription confirmation")
	}
}

func Stop(ctx context.Context, bot *gotgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	logger.Log.Info().Int64("chat_id", chatID).Msg("stop command received")

	res := database.DB.Delete(&model.User{}, "chat_id = ?", chatID)
	switch {
	case res.Error != nil:
		logger.Log.Error().Err(res.Error).Int64("chat_id", chatID).Msg("failed to delete user")
	case res.RowsAffected == 0:
		_, err := bot.SendMessage(ctx, &gotgbot.SendMessageParams{
			ChatID: chatID,
			Text:   "⏹️ Вы не были подписаны на алёрты",
		})
		logger.Log.Debug().Err(err).Int64("chat_id", chatID).Msg("sent not subscribed message")
	default:
		_, err := bot.SendMessage(ctx, &gotgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "❌ Вы отписались от алёртов",
			ReplyMarkup: BuildReplyKeyboard(),
		})
		logger.Log.Debug().Err(err).Int64("chat_id", chatID).Msg("sent unsubscription confirmation")
	}
}

func Status(ctx context.Context, bot *gotgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	logger.Log.Info().Int64("chat_id", chatID).Msg("status command received")

	res := database.DB.First(&model.User{}, "chat_id = ?", chatID)
	switch {
	case res.Error == gorm.ErrRecordNotFound:
		_, err := bot.SendMessage(ctx, &gotgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "❌ Вы не подписаны на алёрты",
			ReplyMarkup: BuildReplyKeyboard(),
		})
		logger.Log.Debug().Err(err).Int64("chat_id", chatID).Msg("sent not subscribed status")
	case res.Error != nil:
		logger.Log.Error().Err(res.Error).Int64("chat_id", chatID).Msg("failed to get user status")
	default:
		_, err := bot.SendMessage(ctx, &gotgbot.SendMessageParams{
			ChatID:      chatID,
			Text:        "✅ Вы подписаны на алёрты",
			ReplyMarkup: BuildReplyKeyboard(),
		})
		logger.Log.Debug().Err(err).Int64("chat_id", chatID).Msg("sent subscribed status")
	}
}

func Help(ctx context.Context, bot *gotgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	logger.Log.Info().Int64("chat_id", chatID).Msg("help command received")

	_, err := bot.SendMessage(ctx, &gotgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        "/start  — ✅ Подписаться на алёрты\n/stop   — ❌ Отписаться от алёртов\n/status — ❔ Проверить статус\n/help   — 📋 Список команд",
		ReplyMarkup: BuildReplyKeyboard(),
	})
	logger.Log.Debug().Err(err).Int64("chat_id", chatID).Msg("sent help message")
}
