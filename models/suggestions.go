package models

type Suggestion struct {
	VideoID, Message, User string
	ID                     int
}

func GetAllSongs() ([]*Suggestion, error) {
	rows, err := db.Query("SELECT videoid, message, user, id FROM submissions;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sugs := make([]*Suggestion, 0)
	for rows.Next() {
		s := new(Suggestion)
		err := rows.Scan(&s.VideoID, &s.Message, &s.User, &s.ID)
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

func AddSuggestion(user string, videoid string, message string) error {
	_, err := db.Exec("INSERT INTO submissions(user, videoid, message) VALUES(?, ?, ?);", user, videoid, message)
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
