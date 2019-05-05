package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

const (
	nick    = "bnflizardwizard"
	channel = "kathleen_lrr"
)

var (
	admins   = [...]string{"kathleen_lrr", "felixphew", "freshpriceofbeleren", "pterodactal"}
	plAdmins = [...]string{"snackpak_", "setralynn"}
)

var (
	ping    = regexp.MustCompile(`^PING :([^\r\n]+)\r\n$`)
	privmsg = regexp.MustCompile(`^:([^\r\n @]+)![^\r\n @]+@[^\r\n @]+\.tmi\.twitch\.tv PRIVMSG [^\r\n ]+ :([^\r\n]+)\r\n$`)
	youtube = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/)([A-Za-z0-9_-]{11})`)
)

var playlist = make(map[string]string)

func irc() {
	for {
		conn, err := tls.Dial("tcp", "irc.chat.twitch.tv:6697", nil)
		if err != nil {
			log.Printf("while connecting: %v; reconnecting in 10 seconds", err)
			time.Sleep(10 * time.Second)
			continue
		}
		_, err = fmt.Fprintf(conn, "PASS oauth:%s\r\n", twitchToken)
		if err != nil {
			log.Printf("while authenticating: %v", err)
			continue
		}
		_, err = fmt.Fprintf(conn, "NICK %s\r\n", nick)
		if err != nil {
			log.Printf("while authenticating: %v", err)
			continue
		}
		_, err = fmt.Fprintf(conn, "JOIN #%s\r\n", channel)
		if err != nil {
			log.Printf("while authenticating: %v", err)
			continue
		}

		r := bufio.NewReader(conn)
		for {
			msg, err := r.ReadString('\n')
			if err != nil {
				log.Printf("while receiving: %v", err)
				break
			}

			if match := ping.FindStringSubmatch(msg); match != nil {
				_, err = fmt.Fprintf(conn, "PONG :%s\r\n", match[1])
				if err != nil {
					log.Printf("while sending: %v", err)
					break
				}
				log.Printf("sent PONG %s", match[1])
			} else if match := privmsg.FindStringSubmatch(msg); match != nil {
				log.Printf("<%s> %s", match[1], match[2])
				err = bot(match[1], match[2], func(msg string) error {
					_, err := fmt.Fprintf(conn, "PRIVMSG #%s :%s\r\n", channel, msg)
					if err == nil {
						log.Printf("> %s", msg)
					}
					return err
				})
				if err != nil {
					log.Printf("while sending: %v", err)
					break
				}
			} else {
				log.Printf("unhandled message: %s", msg)
			}
		}
	}
}

func bot(user, msg string, send func(string) error) (err error) {
	switch {
	case strings.Contains(msg, "!bot"):
		err = send("I Am BNFLizardWizard, A Golem Constructed From My Predecessors By felixphew, " +
			"FreshPrinceOfBeleren, SnackPak_ And Others. I Collect Music Suggestions For Kathleen. " +
			"https://bnf.ffetc.net")
	case strings.Contains(msg, "!howto"):
		err = send("To request a song for Brave New Faves, wait until Kathleen asks for suggestions during the last " +
			"song, then drop a YouTube link in chat, along with the artist's name, song title, and a brief " +
			"description hyping your request.")
	case strings.Contains(msg, "!wiki"):
		err = send("Past playlists can be found on the LoadingReadyWiki: " +
			"https://wiki.loadingreadyrun.com/index.php/Brave_New_Faves")
	case strings.Contains(msg, "!apple"):
		if apple, ok := playlist["apple"]; ok {
			err = send("Tonight's playlist: " + apple)
		}
	case strings.Contains(msg, "!spotify"):
		if spotify, ok := playlist["spotify"]; ok {
			err = send("Tonight's playlist: " + spotify)
		}
	case strings.Contains(msg, "!google"):
		if google, ok := playlist["google"]; ok {
			err = send("Tonight's playlist: " + google)
		}
	case strings.Contains(msg, "!youtube"):
		if youtube, ok := playlist["youtube"]; ok {
			err = send("Tonight's playlist: " + youtube)
		}
	case strings.HasPrefix(msg, "!set_apple"):
		if admin(user, true) {
			playlist["apple"] = msg[len("!set_apple "):]
			err = send("Playlist updated!")
		}
	case strings.HasPrefix(msg, "!set_spotify"):
		if admin(user, true) {
			playlist["spotify"] = msg[len("!set_spotify "):]
			err = send("Playlist updated!")
		}
	case strings.HasPrefix(msg, "!set_google"):
		if admin(user, true) {
			playlist["google"] = msg[len("!set_google "):]
			err = send("Playlist updated!")
		}
	case strings.HasPrefix(msg, "!set_youtube"):
		if admin(user, true) {
			playlist["youtube"] = msg[len("!set_youtube "):]
			err = send("Playlist updated!")
		}
	case strings.HasPrefix(msg, "!clear"):
		if admin(user, false) {
			for p := range playlist {
				delete(playlist, p)
			}
			_, err = db.Exec("DELETE FROM submissions;")
			if err != nil {
				log.Printf("clearing submissions: %v", err)
			}
			err = send("Suggestions and playlists cleared.")
			break
		}
	default:
		if match := youtube.FindStringSubmatch(msg); match != nil {
			if user == "bravenewfavesbot" {
				log.Printf("ignored link: %s", match[1])
			} else {
				log.Printf("found link: %s", match[1])
				_, err := db.Exec("INSERT INTO submissions(user, videoid, message) VALUES(?, ?, ?);",
					user, match[1], msg)
				if err != nil {
					log.Printf("adding submission: %v", err)
				}
			}
		}

	}
	return
}

func admin(user string, playlist bool) bool {
	if playlist {
		for _, u := range plAdmins {
			if user == u {
				return true
			}
		}
	}
	for _, u := range admins {
		if user == u {
			return true
		}
	}
	return false
}
