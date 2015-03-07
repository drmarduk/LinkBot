package main

import (
	"crypto/tls"
	"log"
	"time"

	"github.com/quiteawful/qairc"
)

func StartIrc() {

	ctxIrc = qairc.QAIrc(*cfgNick, *cfgNick)
	ctxIrc.Address = *cfgNetwork + ":" + *cfgPort
	ctxIrc.UseTLS = true
	ctxIrc.TLSCfg = &tls.Config{InsecureSkipVerify: true}

	err := ctxIrc.Run()
	if err != nil {
		log.Println(err.Error())
		return
	}

	for {
		m, status := <-ctxIrc.Out
		if !status {
			ctxIrc.Reconnect()
		}

		if m.Type == "001" {
			ctxIrc.Join(*cfgChannel)
		}
		if m.Type == "PRIVMSG" {
			l := len(m.Args)
			msg := m.Args[l-1]

			p := &Post{
				User:      m.Sender.Nick,
				Message:   msg,
				Timestamp: time.Now(),
			}

			PostReceiver <- p
		}
	}
}
