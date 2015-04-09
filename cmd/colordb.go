package main

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	_ "github.com/drmarduk/go-sqlite3"
	"github.com/pressly/goico"
)

func main() {
	linkcolor := make(map[string]string)
	linkcache := make(map[string]string)
	db, err := sql.Open("sqlite3", "file:../data/links.db?cache=shared&mode=rwc")
	if err != nil {
		log.Println(err.Error())
	}
	res, err := db.Query("SELECT url from links where color IS NULL")
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer res.Close()
	link := ""
	for res.Next() {
		res.Scan(&link)

		u, err := url.Parse(link)
		if err != nil || u.Scheme == "" || u.Host == "" {
			continue
		}
		if _, exists := linkcache[u.Host]; exists {
			linkcolor[link] = linkcache[u.Host]
			continue
		}
		c := GetColor(u)
		log.Println(u.Scheme + "://" + u.Host + ": " + c)
		linkcache[u.Host] = c
		linkcolor[link] = c
	}
	for k, v := range linkcolor {
		db.Exec("UPDATE links set color = ? where url = ?", v, k)
	}
	log.Println("derp")
}

func GetColor(u *url.URL) string {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial:            dialTimeout,
	}
	client := &http.Client{Transport: tr}
	log.Println("next:", u.Scheme+"://"+u.Host+"/favicon.ico")
	resp, err := client.Get(u.Scheme + "://" + u.Host + "/favicon.ico")
	if err != nil {
		log.Println(err.Error())
		return "#FFFFFF"
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	i, err := ico.Decode(bytes.NewReader(buf))
	if err != nil {
		log.Println(err.Error())
		return "#FFFFFF"
	}
	return calcColor(i)
}

func RGBToHex(c color.Color) (h string) {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

func calcColor(img image.Image) string {
	m := make(map[string]int)
	sum := 0
	var maxColor string
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			h := RGBToHex(img.At(x, y))
			m[h]++
			if m[h] > sum {
				maxColor = h
				sum = m[h]
			}
			m[h]++

		}
	}
	return maxColor
}

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, 5*time.Second)
}
