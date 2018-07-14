package main

import (
	"fmt"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

var message = fmt.Sprintf("Hi, I will notify you when a new article will be published on the %s. Here is my commands list.\n\nCommands List:\n/check - check if new article posted\n/subscribe - subscribe for receiving notify when new article posted\n/unsubscribe - unsubscribe from notifications", resource)

// Bot func run telegram bot
func Bot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic("Missing bot token")
	}
	config := tgbotapi.NewUpdate(0)
	updates, _ := bot.GetUpdatesChan(config)
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
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
			bot.Send(msg)
		case "check":
			if Article.Status {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, Article.NewArticle)
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "There are no new articles")
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
