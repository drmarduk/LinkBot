package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thoas/stats"
)

var linksperpage int = 30

func StartApi() {
	middleware := stats.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/get/", getApiHandler)
	mux.HandleFunc("/link/", linkApiHandler)
	mux.HandleFunc("/search/", searchApiHandler)
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(middleware.Data())
		w.Write(b)
	})

	handler := middleware.Handler(mux)
	http.ListenAndServeTLS(*srvAdress+":12345", "data/server.crt", "data/server.key", handler)
}

type Response struct {
	Status int
	Data   []LinkResult
}

type LinkResult struct {
	Id     int       `json:"id"`
	User   string    `json:"user"`
	Url    string    `json:"url"`
	Time   time.Time `json:"time"`
	Post   string    `json:"post,omitempty"`
	Mime   string    `json:"mime,omitempty"`
	Header string    `json:"header,omitempty"`
	Src    string    `json:"src,omitempty"`
}

/*

	- /get			<- default links
	- /get/0		<- default links
	- /get/100		<- default Anzahl ab ID 100 (jünger als ID 100)
	- /get/100/30	<- 30 Links aber ID 100

	- /link/1337	<- Details zu Link mit ID 1337

	- /search/das+ist+ein+suchtext	<- suche nach "das ist ein suchtext" -> pipe to elasticsearch
*/
func getApiHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = strings.Replace(path, "/get/", "", 1)

	args := strings.Split(path, "/")
	var id, count int
	var err error

	if len(args) == 0 || len(args) > 2 {
		id = 0
		count = linksperpage
	} else if len(args) == 1 {
		id, err = strconv.Atoi(args[0])
		if err != nil {
			id = 0
		}
		count = linksperpage
	} else if len(args) == 2 {
		id, err = strconv.Atoi(args[0])
		if err != nil {
			id = 0
		}
		count, err = strconv.Atoi(args[1])
		if err != nil {
			count = linksperpage
		}
	}

	// kk, parsing done.
	db := Db{}
	db.Open()
	query := "select id, user, url, time from links where id < $1 order by id desc limit 0, $2"
	err = db.Prepare(query)
	if err != nil {
		log.Println("Error while preparing Query: " + err.Error())
		db.Close()
		io.WriteString(w, "sorry, error.")
		return
	}
	err = db.QueryStmt(id, count)
	if err != nil {
		log.Println("Error while query: " + err.Error())
		db.Close()
		io.WriteString(w, "error, sorry.")
		return
	}
	defer db.Close() // könnte man auch vorher schon deferen

	var result Response = Response{}
	result.Data = make([]LinkResult, 0)

	for db.ResultRows.Next() {
		var x LinkResult = LinkResult{}
		err = db.ResultRows.Scan(&x.Id, &x.User, &x.Url, &x.Time)
		if err != nil {
			log.Println("Error while scanning row: " + err.Error())
			continue
		}
		result.Data = append(result.Data, x)
	}

	result.Status = 200

	x, _ := json.Marshal(result)
	io.WriteString(w, string(x))
}

func linkApiHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id, err := strconv.Atoi(strings.Replace(path, "/link/", "", 1))
	if err != nil {
		io.WriteString(w, "sorry, error")
		return
	}

	var result Response = Response{}
	result.Data = make([]LinkResult, 0)

	db := Db{}
	db.Open()
	query := "select id, user, url, time, post, mime, header, src from links where id = $1"
	err = db.Prepare(query)
	if err != nil {
		db.Close()
		log.Println("Error while preparing query: " + err.Error())
		io.WriteString(w, "sorry, error")
		return
	}

	err = db.QueryStmt(id)
	if err != nil { // 642058
		db.Close()
		log.Println("Error while query: " + err.Error())
		io.WriteString(w, "sorry, error")
		return
	}

	for db.ResultRows.Next() {
		var id int
		var timestamp time.Time
		var nsUser, nsUrl, nsPost, nsMime, nsHeader, nsSrc sql.NullString

		err = db.ResultRows.Scan(&id, &nsUser, &nsUrl, &timestamp, &nsPost, &nsMime, &nsHeader, &nsSrc)

		if err != nil {
			log.Println("Error while scanning row: " + err.Error())
			continue
		}

		var x LinkResult = LinkResult{}
		x.Id = id
		if x.User = "anon"; nsUser.Valid {
			x.User = nsUser.String
		}
		x.Url = nsUrl.String
		x.Time = timestamp
		if x.Post = ""; nsPost.Valid {
			x.Post = nsPost.String
		}
		if x.Mime = ""; nsMime.Valid {
			x.Mime = nsMime.String
		}
		if x.Header = ""; nsHeader.Valid {
			x.Header = nsHeader.String
		}
		if x.Src = ""; nsSrc.Valid {
			x.Src = nsSrc.String
		}
		result.Data = append(result.Data, x)
	}

	result.Status = 200

	x, _ := json.Marshal(result)
	io.WriteString(w, string(x))
}

func searchApiHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	term := strings.Replace(path, "/search/", "", 1)

	var result Response = Response{}
	result.Data = make([]LinkResult, 0)

	db := Db{}
	db.Open()
	query := "select id, user, url, time from links where src like '%$1%';" //instr(src, $1) > 0"
	err := db.Prepare(query)
	if err != nil {
		db.Close()
		log.Println("Error while preparing query: " + err.Error())
		io.WriteString(w, "sorry, error")
		return
	}

	err = db.QueryStmt(term)
	if err != nil {
		db.Close()
		log.Println("Error while query: " + err.Error())
		io.WriteString(w, "sorry, error")
		return
	}

	for db.ResultRows.Next() {
		var id int
		var timestamp time.Time
		var nsUser, nsUrl sql.NullString

		err = db.ResultRows.Scan(&id, &nsUser, &nsUrl, &timestamp)

		if err != nil {
			log.Println("Error while scanning row: " + err.Error())
			continue
		}

		var x LinkResult = LinkResult{}
		x.Id = id
		if x.User = "anon"; nsUser.Valid {
			x.User = nsUser.String
		}
		x.Url = nsUrl.String
		x.Time = timestamp

		result.Data = append(result.Data, x)
	}

	result.Status = 200

	x, _ := json.Marshal(result)
	io.WriteString(w, string(x))
}
