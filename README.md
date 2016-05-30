# gogitterirc

This is a simple Gitter/IRC/Telegram syncronization bot witten in Go for a low memory footprint.

**NOTICE: It is currently in development, but actively used and maintained!**

## Why?

The [gitter-irc-bot](https://github.com/finnp/gitter-irc-bot) project works great, but the overhead of Node.js is an annoyance so I decided to write a clone in Go (which compiles to native code). Also the Gitter stream API is very unstable so this bot uses the [Gitter IRC bridge](http://irc.gitter.im).

## Installation

1. Install [Go](https://golang.org) (developed on go1.6)
2. Run `go get github.com/mrexodia/gogitterirc`
3. Copy `config_sample.json` to `config.json` and configure like this:
```
{
    "IRC": {
        "Nick": "nickname",
        "Channel": "#channel"
    },
    "Gitter": {
        "Pass": "0123456789abcdef0123456789abcdef01234567",
        "Nick": "nickname",
        "Channel": "#team/room"
    },
    "Telegram": {
        "Token": "012345678:abcdefghijklmn024728734hskjdnchfdb4",
        "Admins": "admin1 admin2 admin3"
    }
}
```
4. Build/Run `gogitterirc` and have fun!

To make the Telegram bot sync run the /startsync command as admin in the group you want to sync to/from.
