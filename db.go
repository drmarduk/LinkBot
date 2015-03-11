package main

import (
	"database/sql"
	"errors"
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

	if db.Stmt != nil {
		db.Stmt.Close()
	}
	db.Stmt = nil

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

func (db *Db) Prepare(query string) error {
	var err error
	db.Stmt = nil
	db.Stmt, err = db.C.Prepare(query)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (db *Db) ExecuteStmt(args ...interface{}) error {
	var err error
	if db.Stmt == nil {
		return errors.New("db.Stmt is nil, use db.Prepare() to create stmt.")
	}

	db.Result, err = db.Stmt.Exec(args...)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func (db *Db) QueryStmt(args ...interface{}) error {
	var err error
	if db.Stmt == nil {
		return errors.New("db.Stmt is nil, use db.Prepare() fist to create stmt.")
	}
	db.ResultRows, err = db.Stmt.Query(args...)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

// Function to install tables
func InstallTables() {
	var tables []string = []string{
		"CREATE TABLE IF NOT EXISTS links(id integer not null primary key, user text, url text, time datetime, post text, mime text, header text, src text);",
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
