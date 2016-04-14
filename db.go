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
	num    int

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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS session (id TEXT PRIMARY KEY, valstr TEXT, valint INTEGER, valtime TIMESTAMP)`)
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

	oldest, err := db.Prepare(`SELECT id, min(login) FROM chars`)
	if err != nil {
		return nil, err
	}

	rem, err := db.Prepare(`DELETE FROM chars WHERE id=?`)
	if err != nil {
		return nil, err
	}

	sadd, err := db.Prepare(`INSERT OR REPLACE INTO session (id, valstr, valint, valtime) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return nil, err
	}

	sget, err := db.Prepare(`SELECT valstr, valint, valtime FROM session WHERE id=?`)
	if err != nil {
		return nil, err
	}

	return &sqliteDB{
		DB: db,

		add:    add,
		get:    get,
		oldest: oldest,
		rem:    rem,

		sadd: sadd,
		sget: sget,
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

func (db *sqliteDB) OldestChar() (id int64, t time.Time, err error) {
	err = db.oldest.QueryRow().Scan(&id, &t)
	return
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
	timeType := reflect.TypeOf(time.Time{})

	err = walkStruct(&s, func(name string, field reflect.Value) error {
		var valstr string
		var valint int64
		var valtime time.Time
		err := db.sget.QueryRow(name).Scan(&valstr, &valint, &valtime)
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
		default:
			if t := field.Type(); timeType.ConvertibleTo(t) {
				field.Set(reflect.ValueOf(valtime).Convert(t))
			}
		}

		return nil
	})

	s.db = db
	return
}

func (db *sqliteDB) SaveSession(s Session) error {
	timeType := reflect.TypeOf(time.Time{})

	return walkStruct(&s, func(name string, field reflect.Value) (err error) {
		switch field.Kind() {
		case reflect.String:
			_, err = db.sadd.Exec(name, field.String(), 0, time.Time{})
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			_, err = db.sadd.Exec(name, "", field.Int(), time.Time{})
		default:
			if field.Type().ConvertibleTo(timeType) {
				_, err = db.sadd.Exec(name, "", 0, field.Convert(timeType).Interface())
			}
		}
		if err != nil {
			log.Fatalf("Failed to save %q (%v): %v", name, field.Interface(), err)
		}

		return
	})
}
