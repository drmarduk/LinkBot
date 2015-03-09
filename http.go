package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func StartHttp() {
	http.HandleFunc("/", homeHandler)
	// jeder Traffic nach https leiten
	go http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 303))
	http.ListenAndServeTLS(*srvAdress+":443", "server.crt", "server.key", nil)
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
		var time interface{} //string //time.Time
		err = rows.Scan(&id, &user, &url, &time)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		io.WriteString(w, fmt.Sprintf("<li>%d - %s <a href='%s'>%s</a> (%v)</li>", id, user, url, url, time))
	}
	io.WriteString(w, "</ul></body></html>")

}
