package main

import (
	"flag"
	"log"

	"github.com/quiteawful/qairc"
)

var (
	ctxIrc       *qairc.Engine
	PostReceiver chan (*Post)

	cfgNick    = flag.String("nick", "Datenkrake", "Nickname")
	cfgUser    = flag.String("user", "marduk", "Username")
	cfgChannel = flag.String("channel", "#rumkugel", "Channel")
	cfgNetwork = flag.String("network", "irc.quiteawful.net", "Network")
	cfgPort    = flag.String("port", "6697", "Ports")
)

func main() {
	flag.Parse()

	InstallTables()
	PostReceiver = make(chan (*Post))

	go StartIrc()
	go StartHttp()
	log.Fatal(StartParser())
}
