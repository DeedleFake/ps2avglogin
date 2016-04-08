package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type DB interface {
	Set(int64, time.Time)
	Get(int64) (time.Time, bool)
	Remove(int64)

	Num() int
	Close() error
}

func createDB() (DB, error) {
	return make(mapDB), nil
}

type mapDB map[int64]time.Time

func (db mapDB) Set(id int64, login time.Time) {
	db[id] = login
}

func (db mapDB) Get(id int64) (time.Time, bool) {
	// This is kind of awkward.
	login, ok := db[id]
	return login, ok
}

func (db mapDB) Remove(id int64) {
	delete(db, id)
}

func (db mapDB) Num() int {
	return len(db)
}

func (db mapDB) Close() error {
	return nil
}

type sqliteDB struct {
	*sql.DB

	add *sql.Stmt
	get *sql.Stmt
	rem *sql.Stmt

	num int
}

func newsqliteDB(path string) (DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`DROP TABLE chars`)
	if err != nil {
		return nil, err
	}

	return &sqliteDB{
		DB: db,
	}, nil
}

func (db *sqliteDB) Set(id int64, login time.Time) {
	panic("Not implemented.")
}

func (db *sqliteDB) Get(id int64) (time.Time, bool) {
	panic("Not implemented.")
}

func (db *sqliteDB) Remove(id int64) {
	panic("Not implemented.")
}

func (db *sqliteDB) Num() int {
	panic("Not implemented.")
}

func (db *sqliteDB) Close() error {
	panic("Not implemented.")
}
