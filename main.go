package main

import (
	"log"

	"github.com/quiteawful/qairc"
)

var (
	ctxIrc       *qairc.Engine
	PostReceiver chan (*Post)
)

func main() {
	InstallTables()
	PostReceiver = make(chan (*Post))

	go StartIrc()
	go StartHttp()
	log.Fatal(StartParser())
}
