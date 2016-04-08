package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type DB interface {
	Set(int64, time.Time) error
	Get(int64) (time.Time, bool, error)
	Remove(int64) error

	Num() int
	Close() error
}

func createDB() (DB, error) {
	switch t := flags.db["type"]; t {
	case "map":
		log.Printf("Using %q for DB.", t)
		return make(mapDB), nil
	case "sqlite", "sqlite3":
		log.Printf("Using %q for DB.", t)
		return newsqliteDB(flags.db["db"])
	}

	return nil, fmt.Errorf("Bad db flag value: %v", flags.db)
}

type mapDB map[int64]time.Time

func (db mapDB) Set(id int64, login time.Time) error {
	db[id] = login
	return nil
}

func (db mapDB) Get(id int64) (time.Time, bool, error) {
	login, ok := db[id]
	return login, ok, nil
}

func (db mapDB) Remove(id int64) error {
	delete(db, id)
	return nil
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

	add, err := db.Prepare(`INSERT OR REPLACE INTO chars (id, login) VALUES (?, ?)`)
	if err != nil {
		return nil, err
	}

	get, err := db.Prepare(`SELECT (login) FROM chars WHERE id=?`)
	if err != nil {
		return nil, err
	}

	rem, err := db.Prepare(`DELETE FROM chars WHERE id=?`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS chars (id INTEGER PRIMARY KEY, login INTEGER)`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`DELETE FROM chars`)
	if err != nil {
		return nil, err
	}

	return &sqliteDB{
		DB: db,

		add: add,
		get: get,
		rem: rem,
	}, nil
}

func (db *sqliteDB) Set(id int64, login time.Time) error {
	_, err := db.add.Exec(id, login)
	if err != nil {
		return err
	}

	db.num++

	return nil
}

func (db *sqliteDB) Get(id int64) (time.Time, bool, error) {
	var login time.Time
	err := db.get.QueryRow(id).Scan(&login)
	switch err {
	case sql.ErrNoRows:
		return login, false, nil
	default:
		return login, false, err
	}

	return login, true, nil
}

func (db *sqliteDB) Remove(id int64) error {
	res, err := db.rem.Exec(id)
	if err != nil {
		return err
	}

	if a, _ := res.RowsAffected(); a > 0 {
		// Should never be more than one, but who knows.
		db.num -= int(a)
	}

	return nil
}

func (db *sqliteDB) Num() int {
	return db.num
}
