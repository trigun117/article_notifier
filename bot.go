package main

import (
	"fmt"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	message1 = fmt.Sprintf("Hi, I'll let you know if a new article is posted on %s. Here is my commands list.\n\nCommands List:\n/check - check if new article posted\n/subscribe - subscribe for notify about new article\n/unsubscribe - unsubscribe from notifications", resource)
	message2 = botMessage
)

// Bot func run telegram bot
func Bot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic("Missing bot token")
	}
	config := tgbotapi.NewUpdate(0)
	updates, err := bot.GetUpdatesChan(config)
	go func() {
		for tick := time.Tick(10 * time.Minute); ; <-tick {
			if Article.Status {
				if err := DataBase.getSubscribers(); err == nil {
					for _, v := range DataBase.Chats {
						msg := tgbotapi.NewMessage(v, Article.NewArticle)
						bot.Send(msg)
					}
				}
				DataBase.Chats = nil
			}
		}
	}()
	for update := range updates {
		switch update.Message.Command() {
		case "start":
			DataBase.addNewUser(update.Message.Chat.UserName, update.Message.Chat.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, message1)
			bot.Send(msg)
		case "check":
			if Article.Status {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, Article.NewArticle)
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, message2)
				bot.Send(msg)
			}
		case "subscribe":
			if result, err := DataBase.subscribe(update.Message.Chat.UserName, update.Message.Chat.ID); err == nil {
				if result == 1 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You have been subscribed")
					bot.Send(msg)
				} else if result == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are already subscribed")
					bot.Send(msg)
				}
			}
		case "unsubscribe":
			if result, err := DataBase.unsubscribe(update.Message.Chat.ID); err == nil {
				if result == 1 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You have been unsubscribed")
					bot.Send(msg)
				} else if result == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You are not subscribed")
					bot.Send(msg)
				}
			}
		}
	}
}
