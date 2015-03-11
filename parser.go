package main

import (
	"fmt"
	"log"
	"regexp"
)

var (
	urlregex *regexp.Regexp = regexp.MustCompile(`((([A-Za-z]{3,9}:(?:\/\/)?)(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;!:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~!#%\/.\w-_]*)?\??(?:[-\+!=&;%@.\w_]*)[#:]?(?:[\w]*))?)`)
)

func StartParser() error {
	for {
		post := <-PostReceiver

		links := extractLink(post.Message)
		for _, l := range links {
			x := Link{
				User:      post.User,
				Url:       l,
				Post:      post.Message,
				Timestamp: post.Timestamp,
			}
			addLink(x)
			log.Printf("%s: %s\n", post.User, l)
		}
	}
}

func extractLink(data string) []string {
	var result []string
	if urlregex.MatchString(data) {
		links := urlregex.FindAllString(data, -1)
		for _, x := range links {
			result = append(result, x)
		}
	}
	return result
}

func addLink(link Link) bool {
	db := Db{}
	db.Open()
	stmt := fmt.Sprintf(`Insert into links(id, user, url, time, post) values(null, "%s", "%s", "%s", "%s")`, link.User, link.Url, link.Timestamp, link.Post)
	err := db.Execute(stmt)
	defer db.Close()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	link.Id, err = db.Result.LastInsertId()

	if err != nil {
		log.Println(err.Error())
		return false
	}
	// TODO: link zum crawler schicken
	CrawlReceiver <- &link
	return true
}
