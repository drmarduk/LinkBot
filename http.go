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
	mux.HandleFunc("/search/", searchHandler)
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(middleware.Data())
		w.Write(b)
	})

	handler := middleware.Handler(mux)
	go http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 301)) // http -> https redirect
	http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", handler)
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

	offset = page * linksperpage

	var links string = ""

	db := Db{}
	db.Open()
	err = db.Query("select id, user, url, time from links order by id desc limit " + strconv.Itoa(offset) + ", " + strconv.Itoa(linksperpage))
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

	// pagination
	var pagination string = "<ul>"
	var total int = totalLinks()

	var totalpages int = int(math.Ceil(float64(total)/float64(linksperpage))) - 1

	switch {
	case page == 0:
		pagination += "<li>0</li>"
		pagination += "<li><a href='/1'>1</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(totalpages) + "'>" + strconv.Itoa(totalpages) + "</a></li>"
		break
	case page == 1:
		pagination += "<li><a href='/'>0</a></li>"
		pagination += "<li>1</li>"
		pagination += "<li><a href='/" + strconv.Itoa(page+1) + "'>" + strconv.Itoa(page+1) + "</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(totalpages) + "'>" + strconv.Itoa(totalpages) + "</a></li>"
		break
	case page > 1 && (page+1) < totalpages:
		pagination += "<li><a href='/'>0</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(page-1) + "'>" + strconv.Itoa(page-1) + "</a></li>"
		pagination += "<li>" + strconv.Itoa(page) + "</li>"
		pagination += "<li><a href='/" + strconv.Itoa(page+1) + "'>" + strconv.Itoa(page+1) + "</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(totalpages) + "'>" + strconv.Itoa(totalpages) + "</a></li>"
		break
	case (page + 1) == totalpages:
		pagination += "<li><a href='/'>0</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(page-1) + "'>" + strconv.Itoa(page-1) + "</a></li>"
		pagination += "<li>" + strconv.Itoa(page) + "</li>"
		pagination += "<li><a href='/" + strconv.Itoa(totalpages) + "'>" + strconv.Itoa(totalpages) + "</a></li>"
		break
	case page == totalpages:
		pagination += "<li><a href='/'>0</a></li>"
		pagination += "<li><a href='/" + strconv.Itoa(page-1) + "'>" + strconv.Itoa(page-1) + "</a></li>"
		pagination += "<li>" + strconv.Itoa(totalpages) + "</li>"
		break
	}

	pagination += "</ul>"
	t := Template{}
	t.Load("index.html")

	t.SetValue("{{lst_Links}}", links)
	t.SetValue("{{lst_Pagination}}", pagination)

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
	var query string = ""
	if für == "" {
		query = "select id, user, url, time from links where instr(lower(post), 'was für') > 0 order by time desc;"
	} else {
		// TODO hello sqli :>
		query = "select id, user, url, time from links where instr(lower(post), 'was für') > 0 and instr(lower(post), lower('" + für + "')) > 0 order by time desc;"
	}

	t := Template{}
	t.Load("index.html")

	var links string = ""

	db := Db{}
	db.Open()
	err := db.Query(query)
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
	io.WriteString(w, t.String())
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

	var term string = strings.Replace(r.URL.Path, "/search/", "", 1)
	var query string = ""
	query = "select id, user, url, time from links where instr(lower(src), lower('" + term + "')) > 0 order by time desc;"

	t := Template{}
	t.Load("index.html")

	var links string = ""

	db := Db{}
	db.Open()
	err := db.Query(query)
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
	io.WriteString(w, t.String())
}

func buildPagintion(current, total int) string {
	return ""
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
