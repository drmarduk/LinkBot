package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func StartHttp2() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/index", homeHandler)
	http.Handle("/", http.FileServer(http.Dir("./html")))
	// jeder Traffic nach https leiten
	go http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 303))
	http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", nil)
}

func StartHttp() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/static/", staticHandler)

	http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", mux)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t := Template{}
	t.Load("index.html")

	var links string = ""

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
		var timestamp time.Time // interface{} //string //time.Time
		err = rows.Scan(&id, &user, &url, &timestamp)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		//links += fmt.Sprintf("<li>%d - %s <a href='%s'>%s</a> (%v)</li>", id, user, url, url, timestamp)
		links += fmt.Sprintf("<li class='lstItem'>%d. <div class='lstUrl'><a href='%s'>%s</a></div><div class='lstMeta'>von %s am %s</div></li>", id, url, url, user, timestamp.Format("02.01.2006 15:04"))

	}

	t.SetValue("{{lst_Links}}", links)
	io.WriteString(w, t.String())
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/"+r.URL.Path[1:])
}
