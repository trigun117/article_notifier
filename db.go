package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

//DB struct
type DB struct {
	Host, Port, User, Password, DBName, SSLMode string
	Chats                                       []int64
}

// DataBase contains data for connect to database
var DataBase = DB{
	Host:     host,
	Port:     port,
	User:     user,
	Password: password,
	DBName:   dbName,
	SSLMode:  sslMode,
}

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
