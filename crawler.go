package main

import (
	"errors"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func StartCrawler() {
	for {
		l := <-CrawlReceiver
		src, err := get(l.Url)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		db := Db{}
		db.Open()
		query := "update links set mime = $1 where id = $2"

		err = db.Prepare(query)
		if err != nil {
			log.Println(err.Error())
            db.Close()
			continue
		}
		err = db.ExecuteStmt(src.MIME, l.Id)
		if err != nil {
			log.Println(err.Error())
            db.Close()
			continue
		}

		query = "insert into search(id, url, src) values($1, $2, $3)"
		err = db.Prepare(query)
		if err != nil {
			log.Println(err.Error())
            db.Close()
			continue
		}
		err = db.ExecuteStmt(l.Id, l.Url, src.Content)
		if err != nil {
			log.Println(err.Error())
            db.Close()
			continue
		}
        db.Close() 
	}
}

//TODO: this does not limit the filesize.
func get(url string) (LinkContent, error) {
	c := LinkContent{}

	out, err := exec.Command("lynx", "-nolist", "-dump", url).CombinedOutput()
	if err != nil {
		log.Println("Crawler.Get:" + err.Error())
		return c, err
	}
	if len(out) == 0 {
		return c, errors.New("Response is empty.")
	}

	c.MIME = http.DetectContentType(out)
	if strings.HasPrefix(c.MIME, "text") {
		c.Content = string(out)
	}
	return c, nil
}
