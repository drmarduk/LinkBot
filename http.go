package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thoas/stats"
)

type Result struct {
	ID        int64
	User      string
	Url       string
	Timestamp time.Time
}

type Pages struct {
	Pagination  []int
	CurrentPage int
}

type HttpResponse struct {
	Results    []Result
	Pagination Pages
}

func StartHttp() {
	middleware := stats.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/static/", staticHandler)
	mux.HandleFunc("/wasfuer/", wasfuerHandler)
	mux.HandleFunc("/search/", searchFormHandler)
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(middleware.Data())
		w.Write(b)
	})

	handler := middleware.Handler(mux)
	go func() {
		log.Fatal(http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 301))) // http -> https redirect
	}()
	log.Fatal(http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	httpRes := HttpResponse{}
	var linksperpage int = 30
	var page int = 0
	var err error
	var offset int = 0
	var x string = strings.Replace(r.URL.Path, "/", "", -1)
	if x != "" {
		page, err = strconv.Atoi(x)
		if err != nil {
			page = 0
		}
	}

	offset = page * linksperpage

	db := Db{}
	db.Open()
	query := "select id, user, url, time from links order by id desc limit $1, $2"
	err = db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		io.WriteString(w, "sorry, error.")
		return
	}
	err = db.QueryStmt(offset, linksperpage)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		io.WriteString(w, "sorry, error")
		return
	}
	defer db.Close()
	for db.ResultRows.Next() {
		res := Result{}
		err = db.ResultRows.Scan(&res.ID, &res.User, &res.Url, &res.Timestamp)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		httpRes.Results = append(httpRes.Results, res)
	}

	// pagination
	var total int = totalLinks()

	var totalpages int = int(math.Ceil(float64(total) / float64(linksperpage)))

	httpRes.Pagination.CurrentPage = page
	httpRes.Pagination.Pagination = buildPagintion(page, totalpages)

	temp, err := template.ParseFiles("html/index.html")
	if err != nil {
		log.Println(err.Error())
	}
	temp.Execute(w, &httpRes)
}

// Handler for static css/js files
func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/"+r.URL.Path[1:])
}

func statsHandler(w http.ResponseWriter, r *http.Request) {

}

func wasfuerHandler(w http.ResponseWriter, r *http.Request) {

	var für string = strings.Replace(r.URL.Path, "/wasfuer/", "", 1)
	var query string = "select id, user, url, time from links where instr(lower(post), 'was für') > 0 and instr(lower(post), lower($1)) > 0 order by time desc;"

	t := Template{}
	t.Load("index.html")

	var links string = ""

	db := Db{}
	db.Open()
	err := db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		io.WriteString(w, "sorry, error.")
		return
	}
	err = db.QueryStmt(für)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		io.WriteString(w, "sorry, error.")
		return
	}
	defer db.Close()

	for db.ResultRows.Next() {
		var id int64
		var user, url string
		var timestamp time.Time
		err = db.ResultRows.Scan(&id, &user, &url, &timestamp)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		links += fmt.Sprintf("<li class='lstItem'>%d. <div class='lstUrl'><a href='%s'>%s</a></div><div class='lstMeta'>von %s am %s</div></li>", id, url, url, user, timestamp.Format("02.01.2006 15:04"))
	}

	t.SetValue("{{lst_Links}}", links)
	t.SetValue("{{lst_Pagination}}", "")
	io.WriteString(w, t.String())
}

func searchFormHandler(w http.ResponseWriter, r *http.Request) {
	term := r.FormValue("term")
	log.Println("Search: " + term)
	var query string = "select id, user, url, time from links where instr(src, $1) > 0 order by time desc;"

	t := Template{}
	t.Load("index.html")

	var links string = ""

	db := Db{}
	db.Open()
	err := db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		io.WriteString(w, "sorry, error.")
		return
	}
	err = db.QueryStmt(term)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		io.WriteString(w, "sorry, error.")
		return
	}
	defer db.Close()

	for db.ResultRows.Next() {
		var id int64
		var user, url string
		var timestamp time.Time
		err = db.ResultRows.Scan(&id, &user, &url, &timestamp)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		links += fmt.Sprintf("<li class='lstItem'>%d. <div class='lstUrl'><a href='%s'>%s</a></div><div class='lstMeta'>von %s am %s</div></li>", id, url, url, user, timestamp.Format("02.01.2006 15:04"))
	}

	t.SetValue("{{lst_Links}}", links)
	t.SetValue("{{lst_Pagination}}", "")
	io.WriteString(w, t.String())
}

func buildPagintion(currentPage, totalPages int) []int {
	var pagination []int
	for i := range iter(totalPages) {

		if i == 0 || i == totalPages-1 || ((i >= currentPage-2) && (i <= currentPage+2)) {
			pagination = append(pagination, i)
		}
	}
	return pagination
}

func totalLinks() int {
	var count int
	db := Db{}
	db.Open()
	err := db.Query("select count(*) as count from links;")
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	db.ResultRows.Next()
	db.ResultRows.Scan(&count)
	db.Close()
	return count
}

func iter(n int) []struct{} {
	return make([]struct{}, n)
}

/*
	Ideen von soda:
		- Links mit "was für" direkt in der Liste mit "von $user am $datum für $user" markieren
		- je nach Mime Typ des Links den Hintergrund des <li> Elements anpassen

	Ideen von svbito:
		- "mach ne anständige json api, faggot" :>

*/
