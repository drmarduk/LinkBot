package main

import (
	"database/sql"
	"log"

	_ "github.com/drmarduk/go-sqlite3"
)

func (db *Db) Open() {
	x, err := sql.Open("sqlite3", "file:data/links.db?loc=CET")
	if err != nil {
		log.Println(err.Error())
	}
	db.C = x
}

func (db *Db) Close() {
	if db.ResultRows != nil {
		db.ResultRows.Close()
	}
	db.Result = nil
	db.C.Close()
}

func (db *Db) Execute(query string) error {
	var err error
	db.Result, err = db.C.Exec(query)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (db *Db) Query(query string) error {
	db.ResultRows = nil
	var err error
	db.ResultRows, err = db.C.Query(query)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

// Function to install tables
func InstallTables() {
	var tables []string = []string{
		"CREATE TABLE IF NOT EXISTS links(id integer not null primary key, user text, url text, time datetime);",
	}
	db := &Db{}
	db.Open()
	for _, s := range tables {
		err := db.Execute(s)
		if err != nil {
			log.Println(err.Error())
		}
	}
	db.Close()
}

func AlterTable() {
	var tables []string = []string{
		"alter table links add column post text;",
		"alter table links add column mime text;",
		"alter table links add column header text;",
		"alter table links add column src text;",
	}

	db := &Db{}
	db.Open()
	for _, s := range tables {
		err := db.Execute(s)
		if err != nil {
			log.Println(err.Error())
		}
	}
	db.Close()

}
