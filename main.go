package main

import (
	"flag"
	"log"

	"github.com/quiteawful/qairc"
)

var (
	ctxIrc       *qairc.Engine
	PostReceiver chan (*Post)
	cfgNick      = flag.String("nick", "Datenkrake2", "Nickname")
	cfgUser      = flag.String("user", "Datenkrake2", "Username")
	cfgChannel   = flag.String("channel", "#rumkugel", "Channel")
	cfgNetwork   = flag.String("network", "irc.quiteawful.net", "Network")
	cfgPort      = flag.String("port", "6697", "Ports")
	srvAdress    = flag.String("host" "links.knilch.net" "host")
)

func main() {
	flag.Parse()

	InstallTables()
	PostReceiver = make(chan (*Post))

	go StartIrc()
	go StartHttp()
	log.Fatal(StartParser())
}
