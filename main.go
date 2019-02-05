package main

import (
	"bnf/models"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"

	_ "github.com/mattn/go-sqlite3"
)

var tmpl *template.Template

func init() {
	models.InitDB()
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/suggest", suggest)
	http.HandleFunc("/post", post)
	http.HandleFunc("/callback", callback)
	http.HandleFunc("/delete", delete)
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	/*if tks == nil {
		http.Redirect(w, r, conf.AuthCodeURL("potato", oauth2.AccessTypeOffline), http.StatusSeeOther)
		return
	}*/
	sugs, err := models.GetAllSongs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "index.html", sugs)
}

func suggest(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "suggest.html", models.Suggestion{})
}

func post(w http.ResponseWriter, r *http.Request) {
	s := models.Suggestion{0,
		r.FormValue("artist"),
		r.FormValue("title"),
		r.FormValue("description"),
		r.FormValue("url"),
		r.FormValue("username"),
	}
	err := models.AddSuggestion(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

func play(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = models.MarkAsPlayed(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = models.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
