package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	urlregex *regexp.Regexp = regexp.MustCompile(`((([A-Za-z]{3,9}:(?:\/\/)?)+(?:[-;:&=\+\$,\w]+@)?[A-Za-z0-9.-]+|(?:www.|[-;!:&=\+\$,\w]+@)[A-Za-z0-9.-]+)((?:\/[\+~!#%\/.\w-_]*)?\??(?:[-\+!=&;%@.\w_]*)[#:]?(?:[\w]*))?)`)
	// Versuch eine leserliche URl regex zu basteln
	// urlregex *regexp.Regexp = regexp.MustCompile(`[a-zA-Z]{3,9}:\/\/((.*)\.)?[a-zA-Z0-9.-]+\.[a-zA-Z]{2,9}(:[0-9]{1,5})?(/(.*)?)?`)
)

var sprueche []string = []string{
	"Obacht! %s hat es am %s schon gepostet.",
	"Aufmerksamkeitsspanne wie ne Fruchtfliege (%s von %s)",
	"AAAALT! (%s von %s)",
	"Dududu! (%s von %s)",
}

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

			// check for duplicate
			result, dup := checkDuplicate(x)
			if dup {

				// wenn *repost* im Post ist, dann nichts sagen
				if !strings.Contains(x.Post, "*repost*") {
					ircMessage(*cfgChannel, fmt.Sprintf(getSpruch(), result.Timestamp.Format("02.01.2006 15:04"), result.User))
				}
				continue
			}

			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}

			resp, err := client.Get(x.Url)
			if err != nil {
				log.Printf("Cannot connect to %s, ignoring: %s\n", x.Url, err.Error())
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

func checkDuplicate(link *Link) (Link, bool) {
	var result Link
	result.User = ""
	db := Db{}
	db.Open()
	defer db.Close()

	err := db.Prepare("Select id, user, url, time from links where url = $1 limit 0, 1")
	if err != nil {
		log.Println("checkDuplicate: %s" + err.Error())
		return result, false
	}

	err = db.QueryStmt(link.Url)
	if err != nil {
		log.Println("checkDuplicate: " + err.Error())
		return result, false
	}

	defer db.ResultRows.Close()
	for db.ResultRows.Next() {
		err = db.ResultRows.Scan(&result.Id, &result.User, &result.Url, &result.Timestamp)
		if err != nil {
			log.Println("checkDuplicate: " + err.Error())
			continue
		}
	}
	log.Printf("%v", result)
	if result.User == "" { // hm, doofer check, besser machen fgt
		return result, false // kein Duplikat
	}
	return result, true // true falls der Link schon in der DB ist, ansonsten false
}

func getSpruch() string {
	return sprueche[rand.Intn(len(sprueche))]
}
