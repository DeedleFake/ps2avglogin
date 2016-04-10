package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"time"
)

type DB interface {
	SetChar(int64, time.Time) error
	GetChar(int64) (time.Time, bool, error)
	RemoveChar(int64) error
	NumChar() int

	LoadSession() (Session, error)
	SaveSession(s Session) error

	Close() error
}

func createDB() (DB, error) {
	switch t := flags.db["type"]; t {
	case "map":
		log.Printf("Using %q for DB.", t)

		if flags.db["s"] == "" {
			flags.db["s"] = "session.json"
		}

		return make(mapDB), nil

	case "sqlite", "sqlite3":
		log.Printf("Using %q for DB.", t)

		return newsqliteDB(flags.db["db"])
	}

	return nil, fmt.Errorf("Bad db flag value: %v", flags.db)
}

type mapDB map[int64]time.Time

func (db mapDB) SetChar(id int64, login time.Time) error {
	db[id] = login
	return nil
}

func (db mapDB) GetChar(id int64) (time.Time, bool, error) {
	login, ok := db[id]
	return login, ok, nil
}

func (db mapDB) RemoveChar(id int64) error {
	delete(db, id)
	return nil
}

func (db mapDB) NumChar() int {
	return len(db)
}

func (db mapDB) LoadSession() (s Session, err error) {
	file, err := os.Open(flags.db["s"])
	if err != nil {
		return s, err
	}
	defer file.Close()

	d := json.NewDecoder(file)
	err = d.Decode(&s)
	s.db = db
	return s, err
}

func (db mapDB) SaveSession(s Session) error {
	file, err := os.Create(flags.db["s"])
	if err != nil {
		return err
	}
	defer file.Close()

	e := json.NewEncoder(file)
	return e.Encode(&s)
}

func (db mapDB) Close() error {
	return nil
}

// TODO: Find a way to do some of this asynchronously.
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

	_, err = db.Exec(`DROP TABLE IF EXISTS chars`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE chars (id INTEGER PRIMARY KEY, login TIMESTAMP)`)
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

	return &sqliteDB{
		DB: db,

		add: add,
		get: get,
		rem: rem,
	}, nil
}

func (db *sqliteDB) SetChar(id int64, login time.Time) error {
	_, err := db.add.Exec(id, login)
	if err != nil {
		return err
	}

	db.num++

	return nil
}

func (db *sqliteDB) GetChar(id int64) (time.Time, bool, error) {
	var login time.Time
	err := db.get.QueryRow(id).Scan(&login)
	if err != nil {
		if err == sql.ErrNoRows {
			return login, false, nil
		}

		return login, false, err
	}

	return login, true, nil
}

func (db *sqliteDB) RemoveChar(id int64) error {
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

func (db *sqliteDB) NumChar() int {
	return db.num
}

func (db *sqliteDB) LoadSession() (s Session, err error) {
	s, err = mapDB{}.LoadSession()
	s.db = db
	return
}

func (db *sqliteDB) SaveSession(s Session) error {
	return mapDB{}.SaveSession(s)
}
