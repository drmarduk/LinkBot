package main

import (
	"crypto/tls"
	"fmt"
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

	// TheMarv will 180 Tage lang einen Timer, wann ich wieder
	// Drachenlord stuff posten darf

	go marv()
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
				Timestamp: time.Now().Local(),
			}

			PostReceiver <- p
		}
	}
}

func ircMessage(channel, msg string) {
	var m string = fmt.Sprintf("PRIVMSG %s :%s\r\n", channel, msg)
	// PRIVMSG #test :Voting time is 600 seconds.
	log.Println(m)
	ctxIrc.In <- m
}

func marv() {
	// start datum
	var startdate time.Time = time.Date(2016, 06, 29, 0, 0, 0, 0, time.UTC)
	var enddate time.Time = startdate.AddDate(0, 0, 180)

	for {
		time.Sleep(1000 * time.Millisecond)

		if time.Now().Hour() == 0 && time.Now().Minute() == 2 {
			ircMessage("rumkugel", "No Drachenlord content until: "+enddate.String())
		}
	}
}
