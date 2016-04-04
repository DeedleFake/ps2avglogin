package main

import (
	"encoding/gob"
	"os"
)

type Session struct {
	Avg RollingAverage
}

func LoadSession(path string) (s Session, err error) {
	file, err := os.Open(path)
	if err != nil {
		return s, err
	}
	defer file.Close()

	d := gob.NewDecoder(file)
	err = d.Decode(&s)
	return s, err
}

func (s Session) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	e := gob.NewEncoder(file)
	return e.Encode(&s)
}
