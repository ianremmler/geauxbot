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
	flipCmd    = "!flip"
	weatherCmd = "!weather"
	table      = "┻━┻"
)

var (
	nick   string
	server string
)

var channel string

func main() {
	log.SetFlags(0)
	log.SetPrefix("igopher: ")

	flag.Usage = func() {
		fmt.Println("usage: igopher [-n <nick>] [-s <server>] #channel")
	}
	flag.StringVar(&nick, "n", "iGopher", "nick of the bot")
	flag.StringVar(&server, "s", "irc.freenode.net", "IRC server")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(0)
	}
	channel = flag.Arg(0)
	if !strings.HasPrefix(channel, "#") {
		log.Fatalf("%s is not a valid channel", channel)
	}

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
		conn.Privmsg(channel, "(ノಠ益ಠ)ノ彡 "+flipped)
	case strings.HasPrefix(text, weatherCmd):
		for _, text := range strings.Split(weather.Info(), "\n") {
			conn.Privmsg(channel, text)
		}
	}
}
