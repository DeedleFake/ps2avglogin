package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"reflect"
	"time"
)

type DB interface {
	SetChar(int64, time.Time) error
	GetChar(int64) (time.Time, bool, error)
	OldestChar() (int64, time.Time, error)
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

		if flags.db["db"] == "" {
			flags.db["db"] = "session.db"
		}

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

func (db mapDB) OldestChar() (int64, time.Time, error) {
	// TODO: Implement.
	panic("Not implemented.")
}

func (db mapDB) RemoveChar(id int64) error {
	delete(db, id)
	return nil
}

func (db mapDB) NumChar() int {
	return len(db)
}

func (db mapDB) LoadSession() (s Session, err error) {
	defer func() {
		s.db = db
	}()

	file, err := os.Open(flags.db["s"])
	if err != nil {
		return s, err
	}
	defer file.Close()

	d := json.NewDecoder(file)
	err = d.Decode(&s)
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

	add    *sql.Stmt
	get    *sql.Stmt
	oldest *sql.Stmt
	rem    *sql.Stmt
	num    *sql.Stmt

	sadd *sql.Stmt
	sget *sql.Stmt
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS session (id TEXT PRIMARY KEY, valstr TEXT, valint INTEGER)`)
	if err != nil {
		return nil, err
	}

	add, err := db.Prepare(`INSERT OR REPLACE INTO chars (id, login) VALUES (?, ?)`)
	if err != nil {
		return nil, err
	}

	get, err := db.Prepare(`SELECT login FROM chars WHERE id=?`)
	if err != nil {
		return nil, err
	}

	oldest, err := db.Prepare(`SELECT id, login, min(login) FROM chars`)
	if err != nil {
		return nil, err
	}

	rem, err := db.Prepare(`DELETE FROM chars WHERE id=?`)
	if err != nil {
		return nil, err
	}

	num, err := db.Prepare(`SELECT count(id) FROM chars`)
	if err != nil {
		return nil, err
	}

	sadd, err := db.Prepare(`INSERT OR REPLACE INTO session (id, valstr, valint) VALUES (?, ?, ?)`)
	if err != nil {
		return nil, err
	}

	sget, err := db.Prepare(`SELECT valstr, valint FROM session WHERE id=?`)
	if err != nil {
		return nil, err
	}

	return &sqliteDB{
		DB: db,

		add:    add,
		get:    get,
		oldest: oldest,
		rem:    rem,
		num:    num,

		sadd: sadd,
		sget: sget,
	}, nil
}

func (db *sqliteDB) SetChar(id int64, login time.Time) error {
	_, err := db.add.Exec(id, login)
	return err
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

func (db *sqliteDB) OldestChar() (id int64, t time.Time, err error) {
	var min string
	err = db.oldest.QueryRow().Scan(&id, &t, &min)
	return
}

func (db *sqliteDB) RemoveChar(id int64) error {
	_, err := db.rem.Exec(id)
	return err
}

func (db *sqliteDB) NumChar() (n int) {
	err := db.num.QueryRow().Scan(&n)
	if err != nil {
		log.Printf("Failed to get number of active characters: %v", err)
	}

	return n
}

func (db *sqliteDB) LoadSession() (s Session, err error) {
	err = walkStruct(&s, func(name string, field reflect.Value) error {
		var valstr string
		var valint int64
		err := db.sget.QueryRow(name).Scan(&valstr, &valint)
		if err != nil {
			if err == sql.ErrNoRows {
				// Just ignore fields that aren't in the database.
				return nil
			}

			return err
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(valstr)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(valint)
		}

		return nil
	})

	s.db = db
	return
}

func (db *sqliteDB) SaveSession(s Session) error {
	return walkStruct(&s, func(name string, field reflect.Value) (err error) {
		switch field.Kind() {
		case reflect.String:
			_, err = db.sadd.Exec(name, field.String(), 0)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			_, err = db.sadd.Exec(name, "", field.Int())
		}
		if err != nil {
			log.Fatalf("Failed to save %q (%v): %v", name, field.Interface(), err)
		}

		return
	})
}
