package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
	"golang.org/x/net/html"
)

// Articles struct
type Articles struct {
	NewArticle, CurrentArticle string
	Status                     bool
}

//DB struct
type DB struct {
	Host, Port, User, Password, DBName, SSLMode string
	Chats                                       []int64
}

// Article contains fresh article
var Article Articles

// DataBase contains data for connect to database
var DataBase = DB{
	Host:     os.Getenv("HOST"),
	Port:     os.Getenv("PORT"),
	User:     os.Getenv("USER"),
	Password: os.Getenv("PASSWORD"),
	DBName:   os.Getenv("DBNAME"),
	SSLMode:  os.Getenv("SSLMODE"),
}

var reg = regexp.MustCompile(`^` + os.Getenv("LINK") + `\d\d\d\d/\d\d/\d\d/.`)

var (
	message1 = os.Getenv("MSG")
	message2 = os.Getenv("MSSG")
)

func (d *DB) addNewUser(username string, chatid int64) error {
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()
	query := `INSERT INTO users(username, chat_id) VALUES($1, $2);`
	if db.Exec(query, `@`+username, chatid); err != nil {
		return err
	}
	return nil
}

func (d *DB) getSubscribers() error {
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT DISTINCT chat_id FROM subscribers;`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		d.Chats = append(d.Chats, id)
	}
	return nil
}

func (d *DB) subscribe(username string, chatid int64) (int64, error) {
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	query := `INSERT INTO subscribers(username, chat_id) SELECT $1, $2 WHERE NOT EXISTS(SELECT * FROM subscribers WHERE chat_id=$3)`
	result, err := db.Exec(query, `@`+username, chatid, chatid)
	if err != nil {
		return 0, err
	}
	r, _ := result.RowsAffected()
	return r, nil
}

func (d *DB) unsubscribe(chatid int64) (int64, error) {
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	query := `DELETE FROM subscribers WHERE chat_id=$1;`
	result, err := db.Exec(query, chatid)
	if err != nil {
		return 0, err
	}
	r, _ := result.RowsAffected()
	return r, nil
}

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
	go func() {
		for tick := time.Tick(10 * time.Minute); ; <-tick {
			if a.Status {
				if err := DataBase.getSubscribers(); err == nil {
					for _, v := range DataBase.Chats {
						msg := tgbotapi.NewMessage(v, a.NewArticle)
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
			if a.Status {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, a.NewArticle)
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

func init() {
	go Article.bot(os.Getenv("TOKEN"))
}

func main() {
	for tick := time.Tick(10 * time.Minute); ; <-tick {
		Article.getCurrentArticle(os.Getenv("LINK"))
		Article.compare()
	}
}
