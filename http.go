package main

import (
	"crypto/tls"
	"encoding/json"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
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
	UrlPrefix   string
	UrlSuffix   string
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
	hwd, err := os.OpenFile(*cfgRoot+"/access.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	middleware = stats.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/robots.txt", staticHandler)
	mux.HandleFunc("/static/", staticHandler)
	mux.HandleFunc("/wasfuer/", wasfuerHandler)
	mux.HandleFunc("/search/", searchFormHandler)
	mux.HandleFunc("/stats", statsHandler)
	mux.HandleFunc("/filter/", filterHandler)
	mux.HandleFunc("/user/", userHandler)

	handler := middleware.Handler(mux)

	go func() {
		log.Fatal(http.ListenAndServe(*srvAdress+":80", http.RedirectHandler("https://"+*srvAdress, 301))) // http -> https redirect
	}()

	config := &tls.Config{MinVersion: tls.VersionTLS10}
	server := http.Server{Addr: *srvAdress + ":443", Handler: WriteLog(handler, hwd), TLSConfig: config}
	log.Fatal(server.ListenAndServeTLS(*cfgRoot+"/data/server.crt", *cfgRoot+"/data/server.key"))
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
	httpRes.Pagination.UrlPrefix = "/"
	renderPage(w, "index.html", &httpRes)
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
	renderPage(w, "index.html", &httpRes)
}

func searchFormHandler(w http.ResponseWriter, r *http.Request) {
	// URL: /search/0/$request string or /search/0/?term=$request string
	var tmp string = strings.Replace(r.URL.Path, "/search/", "", 1)
	httpRes := HttpResponse{}
	var page, total int
	var err error
	var p, t string = "0", ""

	x := strings.Split(tmp, "/")
	if len(x) > 1 {
		p, t = x[0], x[1]
	} else if len(r.URL.RawQuery) > 5 {
		p = x[0]
		t = strings.Replace(r.URL.RawQuery, "term=", "", 1) // no-js
	} else {
		return
	}

	page, err = strconv.Atoi(p)
	if err != nil {
		page = 0
	}

	httpRes.Results, total, err = getSearchLinks(page, t)
	if err != nil {
		log.Println("search: " + err.Error())
	}

	// render
	httpRes.Pagination.TotalPages = total
	httpRes.Pagination.CurrentPage = page
	httpRes.Pagination.UrlPrefix = "/search/"
	httpRes.Pagination.UrlSuffix = "/" + t // might be xss'able?
	renderPage(w, "index.html", &httpRes)
}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	// URL: /filter/0/$filter/$value nojs fuck it
	var tmp string = strings.Replace(r.URL.Path, "/filter/", "", 1)
	httpRes := HttpResponse{}

	var page, total int
	var err error
	var p, f, v string
	x := strings.Split(tmp, "/")
	if len(x) == 3 {
		p, f, v = x[0], x[1], x[2]
	} else {
		// redirect to main page
		return
	}
	page, err = strconv.Atoi(p)
	if err != nil {
		page = 0
	}

	httpRes.Results, total, err = getFilterLinks(page, f, v)
	if err != nil {
		log.Println("filter: " + err.Error())
	}

	httpRes.Pagination.TotalPages = total
	httpRes.Pagination.CurrentPage = page
	httpRes.Pagination.UrlPrefix = "/filter/"
	httpRes.Pagination.UrlSuffix = "/" + f
	renderPage(w, "index.html", &httpRes)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	var tmp string = strings.Replace(r.URL.Path, "/user/", "", 1)
	httpRes := HttpResponse{}

	var page, total int
	var err error
	var p, v string // aktuelle page und username
	x := strings.Split(tmp, "/")
	if len(x) == 2 {
		p, v = x[0], x[1]
	} else {
		return // dann zurück
	}
	page, err = strconv.Atoi(p) // get current page
	if err != nil {
		page = 0
	}

	httpRes.Results, total, err = getFilterLinks(page, "user", v) // todo
	if err != nil {
		log.Println("user: " + err.Error())
	}

	httpRes.Pagination.TotalPages = total
	httpRes.Pagination.CurrentPage = page
	httpRes.Pagination.UrlPrefix = "/user/"
	httpRes.Pagination.UrlSuffix = "/" + v

	renderPage(w, "index.html", &httpRes)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, *cfgRoot+"/html/"+r.URL.Path[1:])
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(middleware.Data())
	w.Write(b)
}

// =============== Link Data retrieving stuff ===============
func getHomeLinks(page int) ([]LinkResult, int, error) {
	links, err := getLinks("order by id desc limit $1, $2;", (page * linksperpage), linksperpage)
	return links, totalPages(";"), err
}

func getWasfürLinks(page int, für string) ([]LinkResult, int, error) {
	links, err := getLinks(
		"where instr(post, 'was für') > 0 and instr(post, $1) > 0 order by id desc limit $2, $3;",
		für, (page * linksperpage), linksperpage)
	return links, totalPages("where instr(post, 'was für') > 0 and instr(post, $1) > 0;", für), err
}

func getSearchLinks(page int, term string) ([]LinkResult, int, error) {
	links, err := getLinks(" join search on links.id = search.id where instr(links.post, $1) > 0 or instr(search.src, $2) > 0 order by links.id desc limit $3, $4;",
		term, term, (page * linksperpage), linksperpage)
	return links, totalPages(" join search on links.id = search.id where instr(links.post, $1) > 0 or instr(search.src, $2) > 0", term, term), err
}

func getFilterLinks(page int, filter, term string) ([]LinkResult, int, error) {
	links, err := getLinks(" join search on links.id = search.id where $1 = $2 order by links.id desc limit $3, $4;", filter, term, (page * linksperpage), linksperpage)
	return links, totalPages(" join search on links.id = search.id where $1 = $2", filter, term), err
}

//func getUserLinks(page int, user, term string) ([]LinkResult, int, error) {
//	links, err := getLinks(" join search on links.id = search.id where user = $1 order by links.id desc limit $2, $3;", user,
//}

func getLinks(query string, args ...interface{}) (result []LinkResult, err error) {
	// mind the order of $1 $2 $3!!! in your query. The matching variables have to be in the same order!!
	result = make([]LinkResult, 0)

	// open Connection
	db := Db{}
	db.Open()
	defer db.Close()

	err = db.Prepare("select links.id, links.user, links.url, links.time from links " + query)
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
	query = "select count(*) from links " + query

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
func renderPage(w http.ResponseWriter, tpl string, result *HttpResponse) {
	result.Pagination.Pagination = buildPagintion(result.Pagination.CurrentPage, result.Pagination.TotalPages)
	temp, err := template.ParseFiles(*cfgRoot + "/html/" + tpl)
	if err != nil {
		log.Println(err.Error())
		return
	}
	temp.Execute(w, result)
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

func iter(n int) []struct{} {
	return make([]struct{}, n)
}
