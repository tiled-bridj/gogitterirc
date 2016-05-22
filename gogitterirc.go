package main

import (
	"fmt"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/configor"
	"github.com/sromku/go-gitter"
	"github.com/thoj/go-ircevent"
)

type Config struct {
	IRC struct {
		Server  string `default:"irc.freenode.net:6667"`
		Nick    string `required:"true"`
		Channel string `required:"true"`
	}
	Gitter struct {
		Token string `required:"true"`
		Room  string `required:"true"`
	}
	Telegram struct {
		Token  string `required:"true"`
		Admins string `required:"true"`
	}
}

func goGitterIrcTelegram(conf Config) {
	//gitter setup
	api := gitter.New(conf.Gitter.Token)
	api.SetDebug(true, nil)

	user, err := api.GetUser()
	if err != nil {
		fmt.Printf("GetUser error: %v\n", err)
		return
	}
	fmt.Printf("[Gitter] Logged in as %v (%v)!\n", user.Username, user.ID)

	room, err := api.JoinRoom(conf.Gitter.Room)
	if err != nil {
		fmt.Printf("JoinRoom error: %v\n", err)
		return
	}
	fmt.Printf("[Gitter] Joined room %v (%v)!\n", room.Name, room.ID)

	api.SendMessage(room.ID, "Hello, I'll be syncronizing between Gitter and IRC/Telegram today!")

	//telegram setup
	bot, err := tgbotapi.NewBotAPI(conf.Telegram.Token)
	if err != nil {
		fmt.Printf("[Telegram] Error in NewBotAPI: %v...\n", err)
		return
	}
	fmt.Printf("[Telegram] Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Printf("[Telegram] Error in GetUpdatesChan: %v...\n", err)
		return
	}

	var groupId int64
	groupId = 0

	//irc setup
	ircCon := irc.IRC(conf.IRC.Nick, conf.IRC.Nick)
	if err := ircCon.Connect(conf.IRC.Server); err != nil {
		fmt.Printf("[IRC] Failed to connect to %v: %v...\n", conf.IRC.Server, err)
		return
	}
	ircCon.AddCallback("001", func(e *irc.Event) {
		ircCon.Join(conf.IRC.Channel)
	})
	ircCon.AddCallback("JOIN", func(e *irc.Event) {
		fmt.Printf("[IRC] Joined channel %v\n", conf.IRC.Channel)
		ircCon.Privmsg(conf.IRC.Channel, "Hello, I'll be syncronizing between IRC and Telegram/Gitter today!")
		ircCon.ClearCallback("JOIN")
	})
	ircCon.AddCallback("PRIVMSG", func(e *irc.Event) {
		gitterMsg := fmt.Sprintf("<%v> %v", e.Nick, e.Message())
		fmt.Printf("[IRC] %v\n", gitterMsg)
		api.SendMessage(room.ID, gitterMsg)
		if groupId != 0 {
			bot.Send(tgbotapi.NewMessage(groupId, gitterMsg))
		}
	})

	//irc loop
	go ircCon.Loop()

	//gitter loop
	stream := api.Stream(room.ID)
	defer stream.Close()
	go api.Listen(stream)

	go func() {
		for {
			event := <-stream.Event
			switch ev := event.Data.(type) {
			case *gitter.MessageReceived:
				if len(ev.Message.From.Username) > 0 && ev.Message.From.Username != user.Username {
					ircMsg := fmt.Sprintf("<%v> %v", ev.Message.From.Username, ev.Message.Text)
					fmt.Printf("[Gitter] %v\n", ircMsg)
					ircCon.Privmsg(conf.IRC.Channel, ircMsg)
					if groupId != 0 {
						bot.Send(tgbotapi.NewMessage(groupId, ircMsg))
					}
				}
			case *gitter.GitterConnectionClosed:
				fmt.Printf("[Gitter] Connection closed...\n")
			}
		}
	}()

	//telegram loop
	for update := range updates {
		message := update.Message
		chat := message.Chat
		name := message.From.UserName
		if len(name) == 0 {
			name = message.From.FirstName
		}
		telegramMsg := fmt.Sprintf("<%s> %s", name, message.Text)
		fmt.Printf("[Telegram] %s\n", telegramMsg)
		if stringInSlice(message.From.UserName, strings.Split(conf.Telegram.Admins, " ")) && strings.HasPrefix(message.Text, "/") {
			if message.Text == "/startsync" && (chat.IsGroup() || chat.IsSuperGroup()) {
				groupId = chat.ID
				bot.Send(tgbotapi.NewMessage(groupId, "Hello, I'll be syncronizing between Telegram and IRC/Gitter today!"))
			} else if message.Text == "/status" {
				bot.Send(tgbotapi.NewMessage(int64(message.From.ID), fmt.Sprintf("Hey! Telegram.groupId: %v, IRC.Connected: %v", groupId, ircCon.Connected())))
			}
		} else if len(message.From.UserName) > 0 && len(message.Text) > 0 {
			if groupId != 0 {
				if groupId != chat.ID { //forward message to group
					bot.Send(tgbotapi.NewMessage(groupId, telegramMsg))
				}
				ircCon.Privmsg(conf.IRC.Channel, telegramMsg)
				api.SendMessage(room.ID, telegramMsg)
			} else {
				fmt.Println("[Telegam] Use /startsync to start the bot...")
			}
		}
	}
}

func gitterTest(conf Config) {
	api := gitter.New(conf.Gitter.Token)
	api.SetDebug(true, nil)

	user, err := api.GetUser()
	if err != nil {
		fmt.Printf("GetUser error: %v\n", err)
		return
	}
	fmt.Printf("[Gitter] Logged in as %v (%v)!\n", user.Username, user.ID)

	room, err := api.JoinRoom(conf.Gitter.Room)
	if err != nil {
		fmt.Printf("JoinRoom error: %v\n", err)
		return
	}
	fmt.Printf("[Gitter] Joined room %v (%v)!\n", room.Name, room.ID)

	api.SendMessage(room.ID, "Hey guys, a bot just joined!")

	faye := api.Faye(room.ID)
	go faye.Listen()

	for {
		event := <-faye.Event
		switch ev := event.Data.(type) {
		case *gitter.MessageReceived:
			if len(ev.Message.From.Username) > 0 && ev.Message.From.Username != user.Username {
				ircMsg := fmt.Sprintf("<%v> %v", ev.Message.From.Username, ev.Message.Text)
				fmt.Printf("[Gitter] %v\n", ircMsg)
			}
		case *gitter.GitterConnectionClosed:
			fmt.Printf("[Gitter] Connection closed...\n")
		}
	}

	/*stream := api.Stream(room.ID)
	defer stream.Close()
	go api.Listen(stream)

	for {
		event := <-stream.Event
		switch ev := event.Data.(type) {
		case *gitter.MessageReceived:
			if ev.Message.From.ID != user.ID {
				ircMsg := fmt.Sprintf("<%v> %v", ev.Message.From.Username, ev.Message.Text)
				fmt.Printf("[Gitter] %v\n", ircMsg)
			}
		case *gitter.GitterConnectionClosed:
			fmt.Printf("[Gitter] Connection closed...\n")
		}
	}*/
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func telegramTest(conf Config) {
	bot, err := tgbotapi.NewBotAPI(conf.Telegram.Token)
	if err != nil {
		fmt.Printf("Error in NewBotAPI: %v...\n", err)
		return
	}
	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Printf("Error in GetUpdatesChan: %v...\n", err)
		return
	}

	var groupId int64
	groupId = 0

	for update := range updates {
		message := update.Message
		chat := message.Chat
		if (chat.IsGroup() || chat.IsSuperGroup()) && stringInSlice(message.From.UserName, strings.Split(conf.Telegram.Admins, " ")) && message.Text == "/startsync" {
			groupId = chat.ID
		} else if len(message.From.UserName) > 0 && len(message.Text) > 0 {
			name := message.From.UserName
			if len(name) == 0 {
				name = message.From.FirstName
			}
			msgText := fmt.Sprintf("%s: %s", name, message.Text)
			fmt.Println(msgText)
			if groupId != 0 {
				if groupId != chat.ID { //forward message to group
					msg := tgbotapi.NewMessage(groupId, msgText)
					bot.Send(msg)
				}
				msg := tgbotapi.NewMessage(groupId, msgText)
				bot.Send(msg)
			} else {
				fmt.Println("Use /startsync to start the bot...")
			}
		}
	}
}

func main() {
	fmt.Println("Gitter/IRC Sync Bot, written in Go by mrexodia")
	var conf Config
	if err := configor.Load(&conf, "config.json"); err != nil {
		fmt.Printf("Error loading config: %v...\n", err)
		return
	}
	goGitterIrcTelegram(conf)
}
