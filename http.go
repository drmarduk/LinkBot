package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thoas/stats"
)

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
	go http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 301)) // http -> https redirect
	log.Fatal(http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
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

	var response Response
	response, err = getLinks(page)

	// pagination
	var pagination string = "<ul class='uk-pagination'>"
	var total int = totalLinks()

	var totalpages int = int(math.Ceil(float64(total)/float64(linksperpage))) - 1

	switch {
	case page == 0:
		pagination += "<li class='uk-active'><span>0</span></li>"
		pagination += "<li><a href='/1'>1</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(totalpages) + "'>" + strconv.Itoa(totalpages) + "</a></li>"
		break
	case page == 1:
		pagination += "<li><a href='/'>0</a></li>"
		pagination += "<li class='uk-active'><span>1</span></li>"
		pagination += "<li><a href='/" + strconv.Itoa(page+1) + "'>" + strconv.Itoa(page+1) + "</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(totalpages) + "'>" + strconv.Itoa(totalpages) + "</a></li>"
		break
	case page > 1 && (page+1) < totalpages:
		pagination += "<li><a href='/'>0</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(page-1) + "'>" + strconv.Itoa(page-1) + "</a></li>"
		pagination += "<li class='uk-active'><span>" + strconv.Itoa(page) + "</span></li>"
		pagination += "<li><a href='/" + strconv.Itoa(page+1) + "'>" + strconv.Itoa(page+1) + "</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(totalpages) + "'>" + strconv.Itoa(totalpages) + "</a></li>"
		break
	case (page + 1) == totalpages:
		pagination += "<li><a href='/'>0</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(page-1) + "'>" + strconv.Itoa(page-1) + "</a></li>"
		pagination += "<li  class='uk-active'><span>" + strconv.Itoa(page) + "</span></li>"
		pagination += "<li><a href='/" + strconv.Itoa(totalpages) + "'>" + strconv.Itoa(totalpages) + "</a></li>"
		break
	case page == totalpages:
		pagination += "<li><a href='/'>0</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(page-1) + "'>" + strconv.Itoa(page-1) + "</a></li>"
		pagination += "<li class='uk-active'><span>" + strconv.Itoa(totalpages) + "</span></li>"
		break
	}

	pagination += "</ul>"
	t := Template{}
	t.Load("index.html")

	t.SetValue("{{lst_Links}}", links)
	t.SetValue("{{lst_Pagination}}", pagination+off)

	io.WriteString(w, t.String())
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

func getLinks(page int) (result Response, err error) {
	// TODO:
	var linksperpage int = 30
	var offset int = page * linksperpage
	result.Data = make([]LinkResult, 0)

	db := Db{}
	db.Open()
	query := "select id, user, url, time from links order by id desc limit $1, $2"
	err = db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		return result, err
	}
	err = db.QueryStmt(offset, linksperpage)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		return result, err
	}
	defer db.Close()

	var id int64
	var user, url string
	var timestamp time.Time

	for db.ResultRows.Next() {
		err = db.ResultRows.Scan(&id, &user, &url, &timestamp)

		if err != nil {
			log.Println(err.Error())
			continue
		}
		result.Data = append(result.Data, LinkResult{Id: int(id), User: user, Url: url, Time: timestamp, TimeStr: timestamp.Format("02.01.2006 15:04")})
	}
	return result, nil
}

/*
	Ideen von soda:
		- Links mit "was für" direkt in der Liste mit "von $user am $datum für $user" markieren
		- je nach Mime Typ des Links den Hintergrund des <li> Elements anpassen

	Ideen von svbito:
		- "mach ne anständige json api, faggot" :>

*/
