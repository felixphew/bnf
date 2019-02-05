package main

import (
	"bnf/models"
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"regexp"
	"time"

	"golang.org/x/oauth2"
)

var tks oauth2.TokenSource

var (
	ping    = regexp.MustCompile(`^PING :([^\r\n]+)\r\n$`)
	privmsg = regexp.MustCompile(`^:([^\r\n @]+)![^\r\n @]+@[^\r\n @]+\.tmi\.twitch\.tv PRIVMSG [^\r\n ]+ :([^\r\n]+)\r\n$`)
	youtube = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/)([A-Za-z0-9_-]{11})`)
)

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
					if match[1] == "bravenewfavesbot" {
						log.Printf("Ignored link %s: <%s> %s", ytmatch[1], match[1], match[2])
					} else {
						log.Printf("Found link %s: <%s> %s", ytmatch[1], match[1], match[2])
						err = models.AddSuggestion(match[1], ytmatch[1], match[2])
						if err != nil {
							log.Print(err)
						}
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
