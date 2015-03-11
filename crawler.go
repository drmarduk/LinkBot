package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

func StartCrawler() {
	for {
		l := <-CrawlReceiver
		src := get(l.Url)

		db := Db{}
		db.Open()
		query := "update links set src = $1 where id = $2"

		err := db.Prepare(query)
		if err != nil {
			log.Println(err.Error())
			db.Close()
			continue
		}
		err = db.ExecuteStmt(src, l.Id)
		if err != nil {
			log.Println(err.Error())
			db.Close()
			continue
		}
		db.Close()
	}
}

func get(url string) string {
	// TODO prevent downloads to be > 10MB or so
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err.Error())
		return ""
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return ""
	}

	src := string(b)
	return src
}
