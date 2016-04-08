package main

import (
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
