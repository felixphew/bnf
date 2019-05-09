package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var tmpl = template.Must(template.ParseGlob("assets/templates/*.html"))

var db *sql.DB

func index(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT videoid, message, user, id FROM submissions;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type sug struct {
		VideoID, Message, User string
		ID                     int
	}
	sugs := []sug{}
	for rows.Next() {
		var s sug
		if err := rows.Scan(&s.VideoID, &s.Message, &s.User, &s.ID); err != nil {
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
	_, err = db.Exec(`INSERT INTO history(user, videoid, message) SELECT user, videoid, message FROM submissions WHERE id = ?1;
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
	rows, err := db.Query("SELECT videoid, message, user, date, id FROM history;")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type sug struct {
		VideoID, Message, User string
		Date                   string
		ID                     int
	}
	sugs := []sug{}
	for rows.Next() {
		var s sug
		if err := rows.Scan(&s.VideoID, &s.Message, &s.User, &s.Date, &s.ID); err != nil {
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

func login(w http.ResponseWriter, r *http.Request) {
	if !auth(r) {
		w.Header().Set("WWW-Authenticate", "Basic realm=bnf")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "bnf.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS submissions(id INTEGER PRIMARY KEY, user TEXT, videoid TEXT, message TEXT);
CREATE TABLE IF NOT EXISTS history (id INTEGER PRIMARY KEY, user TEXT, videoid TEXT, message TEXT, date DATE DEFAULT CURRENT_DATE);
CREATE TABLE IF NOT EXISTS auth(username TEXT, password TEXT);`)
	if err != nil {
		log.Fatal(err)
	}

	go irc()

	http.HandleFunc("/", index)
	http.HandleFunc("/play", play)
	http.HandleFunc("/delete", del)
	http.HandleFunc("/history", history)
	http.HandleFunc("/login", login)
	http.Handle("/bnf.css", http.FileServer(http.Dir("assets")))

	log.Fatal(http.ListenAndServe("127.0.0.1:8001", nil))
}
