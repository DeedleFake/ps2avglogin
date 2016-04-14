package main

import (
	"bytes"
	"time"
)

// A RollingAverage is an average duration that can be updated with
// more data points.
type RollingAverage struct {
	// Cur is the current duration. The jsonDuration type is a light
	// wrapper around time.Duration to make it marshal to and unmarshal
	// from JSON properly.
	Cur jsonDuration `json:"cur"`

	// Num is the number of data points that the current average was
	// calculated from.
	Num int64 `json:"num"`
}

// Update adds a new data point to the average.
func (r *RollingAverage) Update(new time.Duration) time.Duration {
	r.Cur = jsonDuration(((int64(r.Cur) * r.Num) + int64(new)) / (r.Num + 1))
	r.Num++

	return time.Duration(r.Cur)
}

// jsonDuration is a thin wrapper around time.Duration to make it more
// JSON friendly.
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
