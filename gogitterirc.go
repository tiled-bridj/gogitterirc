package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/jinzhu/configor"
	"github.com/thoj/go-ircevent"
)

type Config struct {
	EnableGitter    bool
	EnableDiscourse bool
	IRC struct {
		Server   string `default:"irc.freenode.net:6667"`
		UseTLS   bool   `default:"false"`
		Pass     string `default:""`
		Nick     string `required:"true"`
		Channel  string `required:"true"`
		Identify string `default:""`
	}
	Gitter struct {
		Server  string `default:"irc.gitter.im:6697"`
		Pass    string `required:"true"`
		Nick    string `required:"true"`
		Channel string `required:"true"`
	}
	Discourse struct {
		Server string `default:"https://try.discourse.org"`
	}
}

func gitterEscape(msg string) string {
	// [![asm.png](https://files.gitter.im/x64dbg/x64dbg/0I1c/thumb/asm.png)](https://files.gitter.im/x64dbg/x64dbg/0I1c/asm.png)
	r1 := regexp.MustCompile("^\\[!\\[[^\\]]+\\]\\(https?:\\/\\/files\\.gitter\\.im\\/[^\\/]+\\/[^\\/]+\\/[^\\/]+\\/thumb\\/[^\\)]+\\)\\]\\(([^\\)]+)\\)$")
	msg = r1.ReplaceAllString(msg, "$1")
	// [test.exe](https://files.gitter.im/x64dbg/x64dbg/ROVJ/test.exe)
	r2 := regexp.MustCompile("\\[[^\\]]+\\]\\((https:\\/\\/files\\.gitter\\.im\\/[^\\/]+\\/[^\\/]+\\/[^\\/]+\\/[^\\)]+)\\)$")
	msg = r2.ReplaceAllString(msg, "$1")
	return msg
}

func goGitterIrc(conf Config) {
	//IRC init
	ircCon := irc.IRC(conf.IRC.Nick, conf.IRC.Nick)
	ircCon.UseTLS = conf.IRC.UseTLS
	ircCon.Password = conf.IRC.Pass

	//Gitter init
	gitterCon := irc.IRC(conf.Gitter.Nick, conf.Gitter.Nick)
	gitterCon.UseTLS = true
	gitterCon.Password = conf.Gitter.Pass

	// Discourse init
	discourseCon := DiscourseClient{conf.Discourse.Server}

	// Discourse loop
	if conf.EnableDiscourse {
		discourseLoop := func() {
			lastdate := time.Now().UTC().Format(time.RFC3339)
			for {
				time.Sleep(time.Minute * 15)
				topics, err := discourseCon.FetchTopics(lastdate)
				if err != nil {
					fmt.Printf("[Discourse] Error: %v\n", err)
				}
				if topics != nil {
					lastdate = topics[0].created
					for it := 0 ; it < len(topics) ; it++ {
						gitterMsg := fmt.Sprintf("[Discourse] new Topic: %v - %v", topics[it].title, topics[it].url)
						fmt.Println(gitterMsg)
						ircCon.Privmsg(conf.IRC.Channel, gitterMsg)
					}
				}
			}
		}
		go discourseLoop()
	}

	//Gitter loop
	if conf.EnableGitter {
		if err := gitterCon.Connect(conf.Gitter.Server); err != nil {
			fmt.Printf("[Gitter] Failed to connect to %v: %v...\n", conf.Gitter.Server, err)
			return
		}
		gitterCon.AddCallback("001", func(e *irc.Event) {
			gitterCon.Join(conf.Gitter.Channel)
		})
		gitterCon.AddCallback("JOIN", func(e *irc.Event) {
			//Gitter welcome message
			fmt.Printf("[Gitter] Joined channel %v\n", conf.Gitter.Channel)
			//ignore when other people join
			gitterCon.ClearCallback("JOIN")
		})
		gitterCon.AddCallback("PRIVMSG", func(e *irc.Event) {
			//construct message
			var gitterMsg string
			if e.Nick == "gitter" { //status messages
				gitterMsg = e.Message()
				match, _ := regexp.MatchString("\\[Github\\].+(opened|closed)", gitterMsg) //whitelist
				if !match {
					fmt.Printf("[Gitter Status] %v", gitterMsg)
					return
				}
			} else { //normal messages
				gitterMsg = fmt.Sprintf("<%v> %v", e.Nick, gitterEscape(e.Message()))
			}
			//log message
			fmt.Printf("[Gitter] %v\n", gitterMsg)
			//send to IRC
			ircCon.Privmsg(conf.IRC.Channel, gitterMsg)
		})
		go gitterCon.Loop()
	}

	//IRC loop
	if err := ircCon.Connect(conf.IRC.Server); err != nil {
		fmt.Printf("[IRC] Failed to connect to %v: %v...\n", conf.IRC.Server, err)
		return
	}
	ircCon.AddCallback("001", func(e *irc.Event) {
		if len(conf.IRC.Identify) != 0 {
			ircCon.Privmsg("NickServ", "identify "+conf.IRC.Identify)
			time.Sleep(15 * time.Second)
		}
		ircCon.Join(conf.IRC.Channel)
	})
	ircCon.AddCallback("JOIN", func(e *irc.Event) {
		//IRC welcome message
		fmt.Printf("[IRC] Joined channel %v\n", conf.IRC.Channel)
		//ignore when other people join
		ircCon.ClearCallback("JOIN")
	})
	ircCon.AddCallback("PRIVMSG", func(e *irc.Event) {
		// strip mIRC color codes
		re := regexp.MustCompile("\x1f|\x02|\x03(?:\\d{1,2}(?:,\\d{1,2})?)?")
		msg := re.ReplaceAllString(e.Message(), "")
		//construct/log message
		ircMsg := fmt.Sprintf("`%v` %v", e.Nick, msg)
		fmt.Printf("[IRC] %v\n", ircMsg)
		//send to Gitter
		if conf.EnableGitter {
			gitterCon.Privmsg(conf.Gitter.Channel, ircMsg)
		}
	})
	ircCon.Loop()
}

func main() {
	fmt.Println("Gitter/IRC Proxy Bot")
	var conf Config
	if err := configor.Load(&conf, "config.json"); err != nil {
		fmt.Printf("Error loading config: %v...\n", err)
		return
	}
	goGitterIrc(conf)
}
