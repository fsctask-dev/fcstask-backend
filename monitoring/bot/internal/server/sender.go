package server

import (
	"context"

	"fcstask-monitor-bot/internal/bot"
	"fcstask-monitor-bot/internal/logger"
	user "fcstask-monitor-bot/internal/model"

	gotgbot "github.com/go-telegram/bot"
)

func SendAlertToUsers(bot *bot.Bot, users []user.User, alertText string) {
	go func() {
		for _, user := range users {
			_, err := bot.TgBot.SendMessage(context.Background(), &gotgbot.SendMessageParams{
				ChatID: user.ChatID,
				Text:   alertText,
			})
			if err != nil {
				logger.Log.Error().Err(err).Int64("chat_id", user.ChatID).Msg("failed to send alert")
			}
		}
	}()
}
