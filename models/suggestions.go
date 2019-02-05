package models

type Suggestion struct {
	ID          int
	Title       string
	Artist      string
	Description string
	URL         string
	UserName    string
}

func GetAllSongs() ([]*Suggestion, error) {
	rows, err := db.Query("SELECT title, artist, description, url, username FROM submissions;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sugs := make([]*Suggestion, 0)
	for rows.Next() {
		s := new(Suggestion)
		err := rows.Scan(&s.Title, &s.Artist, &s.Description, &s.URL, &s.UserName)
		if err != nil {
			return nil, err
		}
		sugs = append(sugs, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sugs, nil
}

func AddSuggestion(s Suggestion) error {
	_, err := db.Exec("INSERT INTO submissions(title, artist, description, url, username) VALUES(?, ?, ?, ?, ?);", s.Title, s.Artist, s.Description, s.URL, s.UserName)
	return err
}

func MarkAsPlayed(id int) error {
	_, err := db.Exec(`INSERT INTO history(user, videoid, message, date) SELECT user, videoid, message, CURRENT_DATE FROM submissions WHERE id = ?1;
	DELETE FROM submissions WHERE id = ?1`, id)
	return err
}

func Delete(id int) error {
	_, err := db.Exec("DELETE FROM submissions WHERE id = ?;", id)
	return err
}
