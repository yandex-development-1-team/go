package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func delTgMessage(bot BotAPI, msg *tgbotapi.Message) {
	deleteMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)
	_, _ = bot.Send(deleteMsg)
}
