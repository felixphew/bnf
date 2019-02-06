package models

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "bnf.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS submissions(id INTEGER PRIMARY KEY, title TEXT, artist TEXT, description TEXT, url TEXT, username TEXT);
CREATE TABLE IF NOT EXISTS history (id INTEGER PRIMARY KEY, user TEXT, videoid TEXT, message TEXT, date TEXT);
CREATE TABLE IF NOT EXISTS auth(username TEXT, password TEXT);`)
	if err != nil {
		log.Fatal(err)
	}
}
