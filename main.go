package main

import (
	"flag"
	"log"

	"github.com/quiteawful/qairc"
)

var (
	ctxIrc        *qairc.Engine
	PostReceiver  chan (*Post)
	CrawlReceiver chan (*Link)
	cfgNick       = flag.String("nick", "Datenkrake2", "Nickname")
	cfgUser       = flag.String("user", "Datenkrake2", "Username")
	cfgChannel    = flag.String("channel", "#g0", "Channel")
	cfgNetwork    = flag.String("network", "irc.quiteawful.net", "Network")
	cfgPort       = flag.String("port", "6697", "Ports")
	srvAdress     = flag.String("host", "localhost", "host")
)

func main() {
	flag.Parse()

	InstallTables()
	PostReceiver = make(chan (*Post))
	CrawlReceiver = make(chan (*Link))

	go StartIrc()
	go StartHttp()
	go StartCrawler()
	log.Fatal(StartParser())
}
