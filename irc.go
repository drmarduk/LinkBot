package main

import (
	//"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/quiteawful/qairc"
)

// StartIrc handles all irc messages and sends them to the parser
func StartIrc() {

	ctxIrc = qairc.QAIrc(*cfgNick, *cfgNick)
	ctxIrc.Address = *cfgNetwork + ":" + *cfgPort
	ctxIrc.UseTLS = false
	//ctxIrc.TLSCfg = &tls.Config{InsecureSkipVerify: true}

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

			// change the sender.nick to the fist word, when
			// qabot answers from telegram
			user := m.Sender.Nick
			if m.Sender.Nick == "qabot" || strings.Contains(m.Sender.Nick, "bot") {
				tmp := strings.Split(msg, ":")
				if len(tmp) > 1 {
					user = tmp[0]
				}
			}

			p := &Post{
				User:      user,
				Message:   msg,
				Timestamp: time.Now().Local(),
			}

			postReceiver <- p
		}
	}
}

func ircMessage(channel, msg string) {
	m := fmt.Sprintf("PRIVMSG %s :%s\r\n", channel, msg)
	// PRIVMSG #test :Voting time is 600 seconds.
	log.Println(m)
	ctxIrc.In <- m
}
