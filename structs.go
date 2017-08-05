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
	Id        int64
	User      string
	Url       string
	Post      string
	Timestamp time.Time
	Size      int64
	Content   LinkContent
	Domain    string
}

type LinkContent struct {
	MIME    string
	Content string
}
type Db struct {
	C          *sql.DB
	ResultRows *sql.Rows
	Result     sql.Result
	Stmt       *sql.Stmt
}

type Template struct {
	File    string
	Content string
}
