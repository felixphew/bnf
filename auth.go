package bnf

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func basicAuth(r *http.Request) bool {
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

func twitchAuthz(user string) (bool, bool) {
	var admin bool
	err := db.QueryRow("SELECT admin FROM twitch_authz WHERE user = ?", user).Scan(&admin)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("reading database: %v", err)
		}
		return false, false
	}
	return true, admin
}

func twitchAuthzInsert(user string) error {
	_, err := db.Exec("INSERT INTO twitch_authz(user, admin) VALUES(?, FALSE);", user)
	if err != nil {
		log.Printf("inserting authz: %v", err)
	}
	return err
}

func twitchAuthzDelete(user string) error {
	result, err := db.Exec("DELETE FROM twitch_authz WHERE user = ? AND admin = FALSE;", user)
	if err == nil {
		rows, _ := result.RowsAffected()
		if rows == 0 {
			err = errors.New("no matching non-admin user found")
		}
	}
	if err != nil {
		log.Printf("deleting authz: %v", err)
	}
	return err
}
