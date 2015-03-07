package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func (db *Db) Open() {
	x, err := sql.Open("sqlite3", "links.db")
	if err != nil {
		log.Println(err.Error())
	}
	db.C = x
}

func (db *Db) Close() {
	db.C.Close()
}

func (db *Db) Execute(query string) (sql.Result, error) {
	result, err := db.C.Exec(query)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return result, nil
}

func (db *Db) Query(query string) (*sql.Rows, error) {
	rows, err := db.C.Query(query)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return rows, nil
}

// Function to install tables
func InstallTables() {
	db := &Db{}
	db.Open()
	_, err := db.Execute("create table if not exists links(id integer not null primary key, user text, url text, time datetime);")
	if err != nil {
		log.Println(err.Error())
	}
	defer db.Close()
}
