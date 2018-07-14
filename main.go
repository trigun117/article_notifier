package main

import (
	"os"
	"time"
)

// environment variables
var (
	link     = os.Getenv("LINK")
	resource = os.Getenv("RESOURCE")
	token    = os.Getenv("TOKEN")
	host     = os.Getenv("HOST")
	port     = os.Getenv("PORT")
	user     = os.Getenv("USER")
	password = os.Getenv("PASSWORD")
	dbName   = os.Getenv("DBNAME")
	sslMode  = os.Getenv("SSLMODE")
)

func main() {
	go func() {
		time.Sleep(11 * time.Minute)
		Bot(token)
	}()
	for tick := time.Tick(10 * time.Minute); ; <-tick {
		Article.getCurrentArticle(link)
		Article.compare()
	}
}
