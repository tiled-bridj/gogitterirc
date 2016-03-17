package main

import (
	"fmt"
	"github.com/jinzhu/configor"
	/*"github.com/sromku/go-gitter"*/
	"github.com/thoj/go-ircevent"
)

type Config struct {
	IRC struct {
		Server  string `default:"irc.freenode.net:6667"`
		Nick    string `required:"true"`
		Channel string `required:"true"`
	}
	Gitter struct {
		Apikey string `required:"true"`
		Room   string `required:"true"`
	}
}

func main() {
	fmt.Println("Gitter/IRC Sync Bot, written in Go by mrexodia")
	var conf Config
	if err := configor.Load(&conf, "config.json"); err != nil {
		fmt.Printf("Error loading config: %v...\n", err)
		return
	}

	ircCon := irc.IRC(conf.IRC.Nick, conf.IRC.Nick)
	if err := ircCon.Connect(conf.IRC.Server); err != nil {
		fmt.Printf("Failed to connect to %v: %v...\n", conf.IRC.Server, err)
		return
	}
	ircCon.AddCallback("001", func(e *irc.Event) {
		ircCon.Join(conf.IRC.Channel)
	})
	ircCon.AddCallback("JOIN", func(e *irc.Event) {
		ircCon.Privmsg(conf.IRC.Channel, "Hello, I'll be syncronizing between IRC and Gitter today!")
	})
	ircCon.AddCallback("PRIVMSG", func(e *irc.Event) {
		gitterMsg := fmt.Sprintf("<%v> %v", e.Nick, e.Message())
		fmt.Printf("[IRC] %v\n", gitterMsg)
		ircCon.Privmsg(conf.IRC.Channel, gitterMsg)
	})
	ircCon.Loop()
}
