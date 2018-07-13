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

// environment variables
var (
	link       = os.Getenv("LINK")
	resource   = os.Getenv("RESOURCE")
	botMessage = os.Getenv("MSG")
	token      = os.Getenv("TOKEN")
	host       = os.Getenv("HOST")
	port       = os.Getenv("PORT")
	user       = os.Getenv("USER")
	password   = os.Getenv("PASSWORD")
	dbName     = os.Getenv("DBNAME")
	sslMode    = os.Getenv("SSLMODE")
)

// Article contains fresh article
var Article Articles

// DataBase contains data for connect to database
var DataBase = DB{
	Host:     host,
	Port:     port,
	User:     user,
	Password: password,
	DBName:   dbName,
	SSLMode:  sslMode,
}

var reg = regexp.MustCompile(`^` + link + `\d\d\d\d/\d\d/\d\d/.`)

var (
	message1 = fmt.Sprintf("Hi, I'll let you know if a new article is posted on %s. Here is my commands list.\nCommands List:\n/check - check if new article posted\n/subscribe - subscribe for notify about new article\n/unsubscribe - unsubscribe from notifications", resource)
	message2 = botMessage
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
	go Article.bot(token)
}

func main() {
	for tick := time.Tick(10 * time.Minute); ; <-tick {
		Article.getCurrentArticle(link)
		Article.compare()
	}
}
