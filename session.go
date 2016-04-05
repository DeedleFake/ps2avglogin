package main

import (
	"bytes"
	"encoding/json"
	"os"
	"time"
)

type Session struct {
	Total   RollingAverage `json:"total"`
	NoShort RollingAverage `json:"noshort"`

	Runtime timeDiff `json:"runtime"`
}

func LoadSession(path string) (s Session, err error) {
	file, err := os.Open(path)
	if err != nil {
		return s, err
	}
	defer file.Close()

	d := json.NewDecoder(file)
	err = d.Decode(&s)
	return s, err
}

func (s Session) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	e := json.NewEncoder(file)
	return e.Encode(&s)
}

type timeDiff time.Time

func (t timeDiff) Since() time.Duration {
	return time.Now().Sub(time.Time(t))
}

func (t timeDiff) MarshalJSON() ([]byte, error) {
	str := (t.Since() / time.Minute * time.Minute).String()

	buf := bytes.NewBuffer(make([]byte, 0, len(str)+2))
	buf.WriteByte('"')
	buf.WriteString(str)
	buf.WriteByte('"')

	return buf.Bytes(), nil
}

func (t *timeDiff) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, `"`)
	d, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}

	*t = timeDiff(time.Now().Add(-d))

	return nil
}
