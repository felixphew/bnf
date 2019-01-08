package main

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/oauth2"

	_ "github.com/mattn/go-sqlite3"
)

var tks oauth2.TokenSource

var tmpl = template.Must(template.ParseFiles("index.html"))

var (
	ping    = regexp.MustCompile(`^PING :([^\r\n]+)\r\n$`)
	privmsg = regexp.MustCompile(`^:([^\r\n @]+)![^\r\n @]+@[^\r\n @]+\.tmi\.twitch\.tv PRIVMSG [^\r\n ]+ :([^\r\n]+)\r\n$`)
	youtube = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/)([A-Za-z0-9_-]{11})`)
)

var db *sql.DB

func index(w http.ResponseWriter, r *http.Request) {
	if tks == nil {
		http.Redirect(w, r, conf.AuthCodeURL("potato", oauth2.AccessTypeOffline), http.StatusSeeOther)
		return
	}
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
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, sugs)
}

func callback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	t, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	first := tks == nil
	tks = conf.TokenSource(oauth2.NoContext, t)
	if first {
		go irc()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func delete(w http.ResponseWriter, r *http.Request) {
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

func irc() {
	for {
		conn, err := tls.Dial("tcp", "irc.chat.twitch.tv:6697", nil)
		if err != nil {
			log.Printf("%v (retrying in 10 seconds)", err)
			time.Sleep(10 * time.Second)
			continue
		}
		t, err := tks.Token()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(conn, "PASS oauth:%s\r\n", t.AccessToken)
		fmt.Fprintf(conn, "NICK %s\r\n", "felixphew")
		fmt.Fprintf(conn, "JOIN #%s\r\n", "kathleen_lrr")
		r := bufio.NewReader(conn)
		for {
			msg, err := r.ReadString('\n')
			if err != nil {
				log.Print(err)
				break
			}
			if match := ping.FindStringSubmatch(msg); match != nil {
				fmt.Fprintf(conn, "PONG :%s\r\n", match[1])
				log.Printf("Sent PONG %s", match[1])
			} else if match := privmsg.FindStringSubmatch(msg); match != nil {
				if ytmatch := youtube.FindStringSubmatch(match[2]); ytmatch != nil {
					log.Printf("Found link %s: <%s> %s", ytmatch[1], match[1], match[2])
					_, err := db.Exec("INSERT INTO submissions(user, videoid, message) VALUES(?, ?, ?);", match[1], ytmatch[1], match[2])
					if err != nil {
						log.Print(err)
					}
				} else {
					log.Printf("<%s> %s", match[1], match[2])
				}
			} else {
				log.Printf("Unhandled message: %s", msg)
			}
		}
	}
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "bnf.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS submissions(id INTEGER PRIMARY KEY, user TEXT, videoid TEXT, message TEXT);
CREATE TABLE IF NOT EXISTS auth(username TEXT, password TEXT);`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/callback", callback)
	http.HandleFunc("/delete", delete)
	http.Handle("/bnf.css", http.FileServer(http.Dir(".")))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
