package bnf

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	nick    = "bnflizardwizard"
	channel = "bravenewfaves"
)

var (
	ping    = regexp.MustCompile(`^PING :([^\r\n]+)\r\n$`)
	privmsg = regexp.MustCompile(`^:([^\r\n @]+)![^\r\n @]+@[^\r\n @]+\.tmi\.twitch\.tv PRIVMSG [^\r\n ]+ :([^\r\n]+)\r\n$`)
	link    = regexp.MustCompile(`https?://(?:[a-z0-9-]+\.bandcamp\.com/track/[a-z0-9-]+|(?:(?:www\.)?youtube\.com/watch\?v=|youtu.be/)[A-Za-z0-9_-]{11}|soundcloud.com/[a-z0-9-]+/[a-z0-9-]+)`)

	playlists       = "(apple|spotify|google|youtube)"
	playlistCmds    = regexp.MustCompile("!" + playlists)
	playlistSetCmds = regexp.MustCompile("^!set_" + playlists + " (.*)")

	cmds = regexp.MustCompile("!([^ ]+)")
)

func irc() {
	for {
		conn, err := tls.Dial("tcp", "irc.chat.twitch.tv:6697", nil)
		if err != nil {
			log.Printf("while connecting: %v; reconnecting in 10 seconds", err)
			time.Sleep(10 * time.Second)
			continue
		}
		_, err = fmt.Fprintf(conn, "PASS oauth:%s\r\n", os.Getenv("BNF_TOKEN"))
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
		err = send("Here Are My Instructions: to request a song, wait until Kathleen asks for suggestions " +
			"(just before the last song on the playlist), then drop a YouTube, Bandcamp or Soundcloud link in chat, " +
			"along with the artist's name, song title, and a brief description hyping your request.")
	case strings.Contains(msg, "!wiki"):
		err = send("Past Playlists Can Be Found On The LoadingReadyWiki: " +
			"https://wiki.loadingreadyrun.com/index.php/Brave_New_Faves")
	case strings.Contains(msg, "!vote") || strings.Contains(msg, "!poll"):
		err = send("Vote Lizard! Vote Kathleen! Vote in the Listeners Poll! https://bit.ly/2Wb7Udu")
	case playlistCmds.MatchString(msg):
		match := playlistCmds.FindStringSubmatch(msg)
		if playlist := getPlaylist(match[1]); playlist != nil {
			err = send("Tonight's Playlist: " + *playlist)
		}
	case strings.Contains(msg, "!theme"):
		if theme := getPlaylist("theme"); theme != nil {
			err = send("Current Suggestion Theme: " + *theme)
		}
	case strings.Contains(msg, "!playlist"):
		err = send("Which Playlist Would You Like? (!spotify, !apple, !google, !youtube)")
	case playlistSetCmds.MatchString(msg):
		if admin(user, true) {
			match := playlistSetCmds.FindStringSubmatch(msg)
			err = setPlaylist(match[1], match[2])
			if err == nil {
				err = send("Playlist Updated!")
			}
		}
	case strings.HasPrefix(msg, "!set_theme "):
		if admin(user, true) {
			err = setPlaylist("theme", msg[len("!set_theme "):])
			if err == nil {
				err = send("Theme Updated!")
			}
		}
	case strings.HasPrefix(msg, "!add_cmd "):
		if admin(user, true) {
			parts := strings.SplitN(msg, " ", 3)
			if len(parts) < 3 {
				break
			}
			_, err := db.Exec("INSERT INTO commands(name, value) VALUES(?, ?) "+
				"ON CONFLICT(name) DO UPDATE SET value=excluded.value;", parts[1], parts[2])
			if err == nil {
				err = send("Command !" + parts[1] + " Added!")
			}
		}
	case strings.HasPrefix(msg, "!remove_cmd "):
		if admin(user, true) {
			name := msg[len("!remove_cmd "):]
			_, err := db.Exec("DELETE FROM commands WHERE name = ?;", name)
			if err == nil {
				err = send("Command !" + name + " Removed!")
			}
		}
	case strings.HasPrefix(msg, "!add_user "):
		if admin(user, false) {
			newUser := strings.ToLower(strings.TrimSpace(msg[len("!add_user "):]))
			err = twitchAuthzInsert(newUser)
			if err != nil {
				err = send(fmt.Sprintf("Failed to add user %q", newUser))
			} else {
				err = send(fmt.Sprintf("Successfully authorised user %q", newUser))
			}
		}
	case strings.HasPrefix(msg, "!remove_user "):
		if admin(user, false) {
			newUser := strings.ToLower(strings.TrimSpace(msg[len("!remove_user "):]))
			err = twitchAuthzDelete(newUser)
			if err != nil {
				err = send(fmt.Sprintf("Failed to remove user %q", newUser))
			} else {
				err = send(fmt.Sprintf("Successfully deauthorised user %q", newUser))
			}
		}
	case strings.HasPrefix(msg, "!clear"):
		if admin(user, false) {
			_, err = db.Exec("DELETE FROM submissions; DELETE FROM playlists WHERE name != \"apple\";")
			if err != nil {
				log.Printf("Clearing Submissions: %v", err)
			}
			err = send("Suggestions And Playlists Cleared.")
			break
		}
	case cmds.MatchString(msg):
		var value string
		for _, match := range cmds.FindAllStringSubmatch(msg, -1) {
			err := db.QueryRow("SELECT value FROM commands WHERE name = ?", match[1]).Scan(&value)
			if err != nil {
				if err != sql.ErrNoRows {
					log.Printf("reading database: %v", err)
				}
				continue
			}
			err = send(value)
			break
		}
	default:
		if match := link.FindString(msg); match != "" {
			if user == "bravenewfavesbot" {
				log.Printf("ignored link: %s", match)
			} else {
				log.Printf("found link: %s", match)
				_, err := db.Exec("INSERT INTO submissions(user, link, message) VALUES(?, ?, ?);",
					user, match, msg)
				if err != nil {
					log.Printf("adding submission: %v", err)
				}
			}
		}

	}
	return err
}

func getPlaylist(name string) (value *string) {
	err := db.QueryRow("SELECT value FROM playlists WHERE name = ?", name).Scan(&value)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("reading database: %v", err)
	}
	return
}

func setPlaylist(name string, value string) error {
	_, err := db.Exec("INSERT INTO playlists(name, value) VALUES(?, ?) "+
		"ON CONFLICT(name) DO UPDATE SET value=excluded.value;", name, value)
	return err
}

func admin(user string, playlist bool) bool {
	authz, admin := twitchAuthz(user)
	return authz && (playlist || admin)
}

func init() {
	go irc()
}
