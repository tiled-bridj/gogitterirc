# gogitterirc

[![Gitter](https://badges.gitter.im/mrexodia/gogitterirc.svg)](https://gitter.im/mrexodia/gogitterirc?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [#gogitterirc](https://webchat.freenode.net/?channels=gogitterirc) on Freenode

This is a simple Gitter/IRC syncronization bot witten in Go for a low memory footprint.

**NOTICE: It is currently in development and not considered stable!**

## Why?

The [gitter-irc-bot](https://github.com/finnp/gitter-irc-bot) project works great, but the overhead of Node.js is an annoyance so I decided to write a clone in Go (which compiles to native code).

## Installation

1. Install [Go](https://golang.org) (developed on go1.6)
2. Run `go get github.com/mrexodia/gogitterirc`
3. Copy `config_sample.json` to `config.json` and configure like this:
```
{
    "IRC": {
        "Server": "irc.freenode.net:6667",
        "Nick": "nickname",
        "Channel": "#channel"
    },
    "Gitter": {
        "Apikey": "0123456789abcdef0123456789abcdef01234567",
        "Room": "team/room"
    }
}
```
4. Build/Run `gogitterirc` and have fun!
