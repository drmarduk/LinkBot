package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bradfitz/http2"
)

func StartHttp() {
	var srv http.Server

	//srv.Addr = "localhost:443"

	// register handler
	http.HandleFunc("/", homeHandler)

	http2.ConfigureServer(&srv, &http2.Server{})

	log.Fatal(srv.ListenAndServeTLS("server.crt", "server.key"))

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "<html><head><title>Sammelsurium an Links</title><head><body><ul>")
	db := Db{}
	db.Open()
	rows, err := db.Query("select id, user, url, time from links")
	if err != nil {
		log.Println(err.Error())
		db.Close()
		io.WriteString(w, "sorry, error.")
		return
	}
	defer db.Close()

	for rows.Next() {
		var id int64
		var user, url string
		var time time.Time
		err = rows.Scan(&id, &user, &url, &time)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		io.WriteString(w, fmt.Sprintf("<li>%d - %s <a href='%s'>%s</a> (%q)</li>", id, user, url, url, time))
	}
	io.WriteString(w, "</ul></body></html>")

}
