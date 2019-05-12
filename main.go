package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var tmpl = template.Must(template.ParseGlob("assets/templates/*.html"))

var db *sql.DB

func index(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT link, message, user, id FROM submissions;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type sug struct {
		Link, Message, User string
		ID                  int
	}
	sugs := []sug{}
	for rows.Next() {
		var s sug
		if err := rows.Scan(&s.Link, &s.Message, &s.User, &s.ID); err != nil {
			log.Print(err)
		}
		sugs = append(sugs, s)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "index.html", sugs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func play(w http.ResponseWriter, r *http.Request) {
	if !auth(r) {
		w.Header().Set("WWW-Authenticate", "Basic realm=bnf")
		http.Error(w, "Please login to mark suggestions as played", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = db.Exec(`INSERT INTO history(user, link, message) SELECT user, link, message FROM submissions WHERE id = ?1;
DELETE FROM submissions WHERE id = ?1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func del(w http.ResponseWriter, r *http.Request) {
	if !auth(r) {
		w.Header().Set("WWW-Authenticate", "Basic realm=bnf")
		http.Error(w, "Please login to delete suggestions", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = db.Exec("DELETE FROM submissions WHERE id = ?;", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func history(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT link, message, user, date, id FROM history;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type sug struct {
		Link, Message, User string
		Date                string
		ID                  int
	}
	sugs := []sug{}
	for rows.Next() {
		var s sug
		if err := rows.Scan(&s.Link, &s.Message, &s.User, &s.Date, &s.ID); err != nil {
			log.Print(err)
		}
		sugs = append(sugs, s)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "history.html", sugs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "bnf.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS submissions(id INTEGER PRIMARY KEY, user TEXT, link TEXT, message TEXT);
CREATE TABLE IF NOT EXISTS history (id INTEGER PRIMARY KEY, user TEXT, link TEXT, message TEXT, date DATE DEFAULT CURRENT_DATE);
CREATE TABLE IF NOT EXISTS auth(username TEXT, password TEXT);`)
	if err != nil {
		log.Fatal(err)
	}

	go irc()

	http.HandleFunc("/", index)
	http.HandleFunc("/play", play)
	http.HandleFunc("/delete", del)
	http.HandleFunc("/history", history)
	http.Handle("/bnf.css", http.FileServer(http.Dir("assets")))

	log.Fatal(http.ListenAndServe("127.0.0.1:8001", nil))
}
