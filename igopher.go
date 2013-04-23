package main

import (
	"github.com/fluffle/goirc/client"
	"github.com/ianremmler/flip"
	"github.com/ianremmler/weather"

	"flag"
	"fmt"
	"os"
	"strings"
)

const nick = "iGopher"

const (
	flipCmd = "!flip"
	weatherCmd = "!weather"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}
	if !strings.HasPrefix(os.Args[1], "#") {
		os.Exit(1)
	}
	flag.Parse() // parses the goirc logging flags.

	channel := os.Args[1]
	quit := make(chan bool)
	c := client.SimpleClient(nick)
	c.HandleFunc(client.CONNECTED, func(conn *client.Conn, line *client.Line) {
		conn.Join(channel)
	})
	c.HandleFunc(client.DISCONNECTED, func(conn *client.Conn, line *client.Line) {
		fmt.Println("disconnected :(")
		quit <- true
	})
	c.HandleFunc(client.PRIVMSG, func(conn *client.Conn, line *client.Line) {
		if line.Target() == channel {
			text := strings.TrimSpace(line.Text())
			switch {
			case strings.HasPrefix(text, flipCmd):
				text = strings.TrimSpace(text[len(flipCmd):])
				conn.Action(channel, "(ノಠ益ಠ)ノ彡 "+flip.Flip(text))
			case strings.HasPrefix(text, weatherCmd):
				for _, text := range strings.Split(weather.Info(), "\n") {
					conn.Privmsg(channel, text)
				}
			}
		}
	})
	if err := c.ConnectTo("irc.freenode.net"); err != nil {
		fmt.Printf("Connection error: %s\n", err)
	}
	<-quit
}
