package main

import (
	"bytes"
	"time"
)

type RollingAverage struct {
	Cur jsonDuration `json:"cur"`
	Num int64        `json:"num"`
}

func (r *RollingAverage) Update(new time.Duration) time.Duration {
	r.Cur = jsonDuration(((int64(r.Cur) * r.Num) + int64(new)) / (r.Num + 1))
	r.Num++

	return time.Duration(r.Cur)
}

type jsonDuration time.Duration

func (d jsonDuration) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('"')
	buf.WriteString(time.Duration(d).String())
	buf.WriteByte('"')

	return buf.Bytes(), nil
}

func (d *jsonDuration) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, `"`)
	t, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}

	*d = jsonDuration(t)

	return nil
}
