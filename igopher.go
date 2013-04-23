package main

import (
	"github.com/fluffle/goirc/client"
	"github.com/ianremmler/igopher/flip"
	"github.com/ianremmler/igopher/weather"

	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	nick       = "iGopher"
	flipCmd    = "!flip"
	weatherCmd = "!weather"
	table      = "┻━┻"
)

var channel string

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("usage: igopher #channel")
	}
	if !strings.HasPrefix(os.Args[1], "#") {
		log.Fatalf("usage: %s is not a valid channel", os.Args[1])
	}
	flag.Parse() // parses the goirc logging flags

	channel = os.Args[1]
	quit := make(chan bool)
	c := client.SimpleClient(nick)
	c.HandleFunc(client.CONNECTED, func(conn *client.Conn, line *client.Line) {
		conn.Join(channel)
	})
	c.HandleFunc(client.DISCONNECTED, func(conn *client.Conn, line *client.Line) {
		fmt.Println("disconnected :(")
		quit <- true
	})
	c.HandleFunc(client.PRIVMSG, handlePrivMsg)
	if err := c.ConnectTo("irc.freenode.net"); err != nil {
		fmt.Printf("Connection error: %s\n", err)
	}
	<-quit
}

func handlePrivMsg(conn *client.Conn, line *client.Line) {
	text := strings.TrimSpace(line.Text())
	switch {
	case strings.HasPrefix(text, flipCmd):
		text = strings.TrimSpace(text[len(flipCmd):])
		flipped := ""
		if len(text) > 0 {
			flipped = flip.Flip(text)
		} else {
			flipped = table
		}
		conn.Action(channel, "(ノಠ益ಠ)ノ彡 "+flipped)
	case strings.HasPrefix(text, weatherCmd):
		for _, text := range strings.Split(weather.Info(), "\n") {
			conn.Privmsg(channel, text)
		}
	}
}
