package main

import (
	"database/sql"
	"time"
)

type Post struct {
	User      string
	Message   string
	Timestamp time.Time
}

type Link struct {
	User      string
	Url       string
	Timestamp time.Time
}

type Db struct {
	C *sql.DB
}

type Template struct {
	File    string
	Content string
}
