package main

import (
	"database/sql"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func auth(r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}

	hash := make([]byte, 0, 60)
	err := db.QueryRow("SELECT password FROM auth WHERE username = ?", user).Scan(&hash)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("reading database: %v", err)
		}
		return false
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(pass))
	if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
		log.Printf("checking password: %v", err)
	}
	return err == nil
}
