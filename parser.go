package main

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
)

var (
	urlregex *regexp.Regexp = regexp.MustCompile(`((([A-Za-z]{3,9}:(?:\/\/)?)+(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;!:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~!#%\/.\w-_]*)?\??(?:[-\+!=&;%@.\w_]*)[#:]?(?:[\w]*))?)`)
	// Versuch eine leserliche URl regex zu basteln
	// urlregex *regexp.Regexp = regexp.MustCompile(`[a-zA-Z]{3,9}:\/\/((.*)\.)?[a-zA-Z0-9.-]+\.[a-zA-Z]{2,9}(:[0-9]{1,5})?(/(.*)?)?`)
)

func StartParser() error {
	for {
		post := <-PostReceiver

		links := extractLink(post.Message)
		for _, l := range links {
			x := &Link{User: post.User, Url: l, Post: post.Message, Timestamp: post.Timestamp}

			u, err := url.Parse(x.Url)
			if err != nil {
				log.Println("unable to parse URL", x.Url)
				continue
			}
			//assuming a sane default
			if u.Scheme == "" {
				x.Url = "http://" + x.Url
			}
			resp, err := http.Get(x.Url)
			if err != nil {
				log.Printf("Cannot connect to %s, ignoring", x.Url)
				continue
			}
			resp.Body.Close()

			if addLink(x) {
				log.Printf("%s: %s\n", post.User, x.Url)
			} else {
				log.Printf("Could not insert link (%s) into the database.\n", x.Url)
			}

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

func addLink(link *Link) bool {
	db := Db{}
	db.Open()
	defer db.Close()

	err := db.Prepare("Insert into links(user, url, time, post) values($1, $2, $3, $4)")
	if err != nil {
		log.Println("addLink: " + err.Error())
		return false
	}

	err = db.ExecuteStmt(link.User, link.Url, link.Timestamp, link.Post)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	link.Id, err = db.Result.LastInsertId()
	if err != nil {
		log.Println(err.Error())
		return false
	}

	CrawlReceiver <- link
	return true
}
