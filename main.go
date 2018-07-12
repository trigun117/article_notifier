package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/html"
)

// Articles struct contains articles
type Articles struct {
	NewArticle, CurrentArticle string
	Status                     bool
}

// Article contain fresh article
var Article Articles

var reg = regexp.MustCompile(os.Getenv("REG"))

var message1, message2 = os.Getenv("MSG"), os.Getenv("MSSG")

func (a *Articles) getCurrentArticle(link string) error {
	response, err := http.Get(link)
	if err != nil {
		return err
	}
	defer func() {
		defer response.Body.Close()
		ioutil.ReadAll(response.Body)
		io.Copy(ioutil.Discard, response.Body)
	}()
	z := html.NewTokenizer(response.Body)
	for {
		switch token := z.Next(); token {
		case html.ErrorToken:
			return nil
		case html.StartTagToken:
			t := z.Token()
			isAnchor := t.Data == "a"
			if isAnchor {
				for _, v := range t.Attr {
					if v.Key == "href" {
						if reg.MatchString(v.Val) {
							a.CurrentArticle = v.Val
							return nil
						}
					}
				}
			}
		}
	}

}

func (a *Articles) compare() {
	if a.CurrentArticle != a.NewArticle {
		a.NewArticle = a.CurrentArticle
		a.Status = true
	} else {
		a.Status = false
	}
}

func (a *Articles) bot(token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic("Missing bot token")
	}
	config := tgbotapi.NewUpdate(0)
	updates, err := bot.GetUpdatesChan(config)
	for update := range updates {
		if update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, message1)
			bot.Send(msg)
		} else if update.Message.Command() == "check" {
			if a.Status {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, a.NewArticle)
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, message2)
				bot.Send(msg)
			}
		}
	}
}

func init() {
	go Article.bot(os.Getenv("TOKEN"))
}

func main() {
	for tick := time.Tick(10 * time.Minute); ; <-tick {
		Article.getCurrentArticle(os.Getenv("LINK"))
		Article.compare()
	}
}
