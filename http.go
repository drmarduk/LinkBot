package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thoas/stats"
)

type LinkResult struct {
	ID        int64
	User      string
	Url       string
	Timestamp time.Time
	TimeStr   string
}

type Pages struct {
	Pagination  []int
	CurrentPage int
}

type HttpResponse struct {
	Results    []LinkResult
	Pagination Pages
}

var (
	linksperpage int = 30
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

	go func() {
		log.Fatal(http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 301))) // http -> https redirect
	}()
	log.Fatal(http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	httpRes := HttpResponse{}
	var page int = 0
	var err error
	var x string = strings.Replace(r.URL.Path, "/", "", -1)
	if x != "" {
		page, err = strconv.Atoi(x)
		if err != nil {
			page = 0
		}
	}

	//var response Response

	httpRes.Results, err = getLinks(0, page, "")
	if err != nil {
		log.Println("homeHandler: " + err.Error())
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
	httpRes := HttpResponse{}
	var page int = 0
	var err error
	//var offset int = 0
	var x string = strings.Replace(r.URL.Path, "/", "", -1)
	if x != "" {
		page, err = strconv.Atoi(x)
		if err != nil {
			page = 0
		}
	}

	httpRes.Results, err = getLinks(1, page, für)
	if err != nil {
		log.Println("wasfuerHandler: " + err.Error())
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

func searchFormHandler(w http.ResponseWriter, r *http.Request) {
	var tmp string = strings.Replace(r.URL.Path, "/search/", "", 1)
	httpRes := HttpResponse{}
	var page int = 0
	var err error
	var p, t string = "0", ""

	x := strings.Split(tmp, "/")
	if len(x) > 1 {
		p, t = x[0], x[1]
	} else {
		return // ordentlich abbrechen, wenn falsche Anzahl an parametern gegeben ist
	}

	page, err = strconv.Atoi(p)
	if err != nil {
		page = 0
	}

	httpRes.Results, err = getLinks(2, page, t)
	if err != nil {
		log.Println("search: " + err.Error())
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

// getLinks returns an arry of LinkResult, it is a wrappper to cover "all types
// of link requests.
//
// typ specifies the typ of handler the links are for. 0 is the basic home-handler
// the where var can be omitted. 1 is for the wasfuer handler, the where var must
// be user. And 2 is for a general search and the where var must be a searchterm.
//
// The page var is used for the "limit 0, x" stuff for the pagination and sets the
// current page.
func getLinks(typ, page int, where string) (result []LinkResult, err error) {
	var offset int
	result = make([]LinkResult, 0)

	if page < 0 { // prevent negative pagecounts
		page = 0
	}
	offset = page * linksperpage

	// create query
	query := "select id, user, url, time from links "

	switch typ {
	case 0:
		query += " order by id desc limit $1, $2;"
		break
	case 1: // wasfuer
		if query += "where instr(post, 'was für') > 0 and instr(post, $1) > 0 order by id desc limit $2, $3;"; where == "" {
			return result, errors.New("where variable must be set when using typ 1")
		}
		break
	case 2: // normal search
		if query += "where instr(src, $1) > 0 order by id desc limit $2, $3"; where == "" {
			return result, errors.New("where variable must be set when using typ 2")
		}
	}

	// open Connection
	db := Db{}
	db.Open()

	err = db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		db.Close()
		return result, err
	}

	switch typ {
	case 0:
		err = db.QueryStmt(offset, linksperpage)
		break
	case 1:
		err = db.QueryStmt(where, offset, linksperpage)
		break
	case 2:
		// will be changed when the source content is served from another
		// table
		err = db.QueryStmt(where, offset, linksperpage)
		break
	}
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
		result = append(result, LinkResult{ID: id, User: user, Url: url, Timestamp: timestamp, TimeStr: timestamp.Format("02.01.2006 15:04")})
	}
	return result, nil
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
