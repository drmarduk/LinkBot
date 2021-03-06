package main

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func (db *Db) Open() {
	x, err := sql.Open("sqlite3", "file:"+*cfgRoot+"/data/links.db")
	if err != nil {
		// ze fuck i did here o.O
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

// func (db *Db) QueryRow(queryVjj string) error {
// 	db.Result = nil
// 	var err error
// 	db.ResultRows = db.C.QueryRow(query)
// 	if err != nil {
// 		log.Printf("error while QueryRow(%s): %v\n", query, err)
// 		return err
// 	}
// 	return nil
// }

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
		"CREATE TABLE IF NOT EXISTS links(id integer not null, user text, url text, host text, time datetime, post text, mime text, primary key(id));",
		"CREATE VIRTUAL TABLE IF NOT EXISTS search USING fts4(id, url, src);",
		"CREATE INDEX IF NOT EXISTS user ON links (user ASC);",
		"CREATE INDEX IF NOT EXISTS domain ON links (domain ASC);",
		"CREATE INDEX IF NOT EXISTS mime ON links (mime ASC);",
	}
	db := &Db{}
	db.Open()
	for _, s := range tables {
		err := db.Execute(s)
		if err != nil {
			log.Println(err.Error())
			panic(err)
		}
	}
	db.Close()
}
