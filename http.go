package main

import (
	"encoding/json"
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
	TotalPages  int
}

type HttpResponse struct {
	ShowError    bool
	ErrorMessage string
	Results      []LinkResult
	Pagination   Pages
}

var (
	linksperpage int = 30
	middleware   *stats.Stats
)

func StartHttp() {
	middleware = stats.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/static/", staticHandler)
	mux.HandleFunc("/wasfuer/", wasfuerHandler)
	mux.HandleFunc("/search/", searchFormHandler)
	mux.HandleFunc("/stats", statsHandler)

	handler := middleware.Handler(mux)

	go func() {
		log.Fatal(http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 301))) // http -> https redirect
	}()
	log.Fatal(http.ListenAndServeTLS(*srvAdress+":443", "data/server.crt", "data/server.key", handler))
}

// =============== Handler ===============
func homeHandler(w http.ResponseWriter, r *http.Request) {
	httpRes := HttpResponse{}
	var page, total int
	var err error
	var x string = strings.Replace(r.URL.Path, "/", "", -1)
	if x != "" {
		page, err = strconv.Atoi(x)
		if err != nil {
			page = 0
		}
	}
	// get data from db
	httpRes.Results, total, err = getHomeLinks(page) // returns links, totalpages and error
	if err != nil {
		log.Println("homeHandler: " + err.Error())
	}

	// pagination
	httpRes.Pagination.TotalPages = total
	httpRes.Pagination.CurrentPage = page
	httpRes.Pagination.Pagination = buildPagintion(page, httpRes.Pagination.TotalPages)

	temp, err := template.ParseFiles("html/index.html")
	if err != nil {
		log.Println(err.Error())
	}
	temp.Execute(w, &httpRes)
}

func wasfuerHandler(w http.ResponseWriter, r *http.Request) {
	var für string = strings.Replace(r.URL.Path, "/wasfuer/", "", 1)
	httpRes := HttpResponse{}
	var page, total int
	var err error
	var x string = strings.Replace(r.URL.Path, "/", "", -1)
	if x != "" {
		page, err = strconv.Atoi(x)
		if err != nil {
			page = 0
		}
	}

	httpRes.Results, total, err = getWasfürLinks(page, für)
	if err != nil {
		log.Println("wasfuerHandler: " + err.Error())
	}

	// pagination
	httpRes.Pagination.TotalPages = total
	httpRes.Pagination.CurrentPage = page
	httpRes.Pagination.Pagination = buildPagintion(page, httpRes.Pagination.TotalPages)

	temp, err := template.ParseFiles("html/index.html")
	if err != nil {
		log.Println(err.Error())
	}
	temp.Execute(w, &httpRes)
}

func searchFormHandler(w http.ResponseWriter, r *http.Request) {
	var tmp string = strings.Replace(r.URL.Path, "/search/", "", 1)
	httpRes := HttpResponse{}
	var page, total int
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

	httpRes.Results, total, err = getSearchLinks(page, t)
	if err != nil {
		log.Println("search: " + err.Error())
	}

	// pagination
	httpRes.Pagination.TotalPages = total
	httpRes.Pagination.CurrentPage = page
	httpRes.Pagination.Pagination = buildPagintion(page, httpRes.Pagination.TotalPages)

	temp, err := template.ParseFiles("html/index.html")
	if err != nil {
		log.Println(err.Error())
	}
	temp.Execute(w, &httpRes)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/"+r.URL.Path[1:])
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(middleware.Data())
	w.Write(b)
}

// =============== Data retrieving stuff ===============
func getHomeLinks(page int) ([]LinkResult, int, error) {
	links, err := getLinks("select id, user, url, time from links order by id desc limit $1, $2;", (page * linksperpage), linksperpage)
	return links, totalPages("select count(*) from links;"), err
}

func getWasfürLinks(page int, für string) ([]LinkResult, int, error) {
	links, err := getLinks(
		"select id, user, url, time from links where instr(post, 'was für') > 0 and instr(post, $1) > 0 order by id desc limit $2, $3;",
		für, (page * linksperpage), linksperpage)
	return links, totalPages("select count(*) from links where instr(post, 'was für') > 0 and instr(post, $1) > 0 order by id desc limit $2, $3;", für, (page * linksperpage), linksperpage), err
}

func getSearchLinks(page int, term string) ([]LinkResult, int, error) {
	links, err := getLinks("select id, user, url, time from links where instr(post, $1) > 0 order by id desc limit $2, $3;",
		term, (page * linksperpage), linksperpage)
	return links, totalPages("select count(*) from links where instr(post, $1) > 0 order by id desc limit $2, $3;", term, (page * linksperpage), linksperpage), err
}

func getLinks(query string, args ...interface{}) (result []LinkResult, err error) {
	// mind the order of $1 $2 $3!!! in your query. The matching variables have to be in the same order!!
	result = make([]LinkResult, 0)

	// open Connection
	db := Db{}
	db.Open()
	defer db.Close()

	err = db.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		return result, err
	}

	err = db.QueryStmt(args...)
	if err != nil {
		log.Println(err.Error())
		return result, err
	}

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

// =============== Data helper functions ===============
func totalLinks(query string, args ...interface{}) int {
	var count int
	var err error
	db := Db{}
	db.Open()

	if args == nil {
		err = db.Query(query)
	} else {
		err = db.Prepare(query)
		if err != nil {
			log.Println(err.Error())
			return 0
		}
		err = db.QueryStmt(args...)
	}

	if err != nil {
		log.Println(err.Error())
		return 0
	}
	db.ResultRows.Next()
	db.ResultRows.Scan(&count)
	db.Close()
	return count
}

func totalPages(query string, args ...interface{}) int {
	return int(math.Ceil(float64(totalLinks(query, args...)) / float64(linksperpage)))
}

// =============== render HTML page functions ===============
func buildPagintion(currentPage, totalPages int) []int {
	var pagination []int
	for i := range iter(totalPages) {

		if i == 0 || i == totalPages-1 || ((i >= currentPage-2) && (i <= currentPage+2)) {
			pagination = append(pagination, i)
		}
	}
	return pagination
}

func iter(n int) []struct{} {
	return make([]struct{}, n)
}
