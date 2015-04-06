package main

import (
	"fmt"
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
		defer db.Close()
		query := "update links set src = $1 where id = $2"

		err = db.Prepare(query)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		err = db.ExecuteStmt(src.Content, l.Id)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}
}

//TODO: this does not limit the filesize.
func get(url string) (LinkContent, error) {
	c := LinkContent{}
	out, err := exec.Command("lynx", "--dump", "-nolist", url).CombinedOutput()
	if err != nil {
		return c, err
	}
	if len(out) == 0 {
		return c, fmt.Errorf("%s: empty response")
	}
	c.MIME = http.DetectContentType(out)
	if strings.HasPrefix(c.MIME, "text") {
		c.Content = string(out)
	}
	return c, nil
}
