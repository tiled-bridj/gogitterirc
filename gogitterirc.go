package main

import (
	"fmt"
	/* "github.com/sromku/go-gitter" */
	"github.com/thoj/go-ircevent"
)

func main() {
	fmt.Println("Gitter/IRC Sync Bot, written in Go by mrexodia")

	ircNick := "gogitterirc"
	ircServ := "irc.freenode.net:6667"
	ircChan := "#x64dbg"
	ircCon := irc.IRC(ircNick, ircNick)
	ircErr := ircCon.Connect(ircServ)
	if ircErr != nil {
		fmt.Printf("Failed to connect to %v as %v", ircServ, ircNick)
		return
	}
	ircCon.AddCallback("001", func(e *irc.Event) {
		ircCon.Join(ircChan)
	})
	ircCon.AddCallback("JOIN", func(e *irc.Event) {
		ircCon.Privmsg(ircChan, "Hello, I'll be syncronizing between IRC and Gitter today!")
	})
	ircCon.AddCallback("PRIVMSG", func(e *irc.Event) {
		gitterMsg := fmt.Sprintf("<%v> %v", e.Nick, e.Message())
		fmt.Printf("[IRC] %v", gitterMsg)
		ircCon.Privmsg(ircChan, gitterMsg)
	})
	ircCon.Loop()
}
