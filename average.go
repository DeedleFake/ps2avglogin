package main

import (
	"bytes"
	"strconv"
	"time"
)

type RollingAverage struct {
	Cur time.Duration
	Num int64
}

func (r *RollingAverage) Update(new time.Duration) time.Duration {
	r.Cur = ((r.Cur * time.Duration(r.Num)) + new) / (time.Duration(r.Num) + 1)
	r.Num++

	return r.Cur
}

func (r RollingAverage) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`{"cur": "`)
	buf.WriteString(r.Cur.String())
	buf.WriteString(`", "num": `)
	buf.WriteString(strconv.FormatInt(r.Num, 10))
	buf.WriteByte('}')

	return buf.Bytes(), nil
}
