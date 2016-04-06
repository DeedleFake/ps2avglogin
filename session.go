package main

import (
	"bytes"
	"encoding/json"
	"os"
	"time"
)

// A Session is the current monitoring session. It keeps track of the
// averages, how long the tracker has been running, etc.
type Session struct {
	// Total is the total average of all session.
	Total RollingAverage `json:"total"`

	// NoShort is an average that excludes 'short' sessions. See
	// flags.short.
	NoShort RollingAverage `json:"noshort"`

	// Longest and Shortest are the longest and shortest sessions that
	// have completed this session, respectively.
	Longest  jsonDuration `json:"longest"`
	Shortest jsonDuration `json:"shortest"`

	// TODO: Add another average that doesn't include repeat characters?

	// Runtime is a timestamp of the time that the tracker was started.
	// timeDiff is a wrapper around time.Time.
	//
	// TODO: Make this a slice that can track the lengths of multiple
	// uses of the same session.
	Runtime timeDiff `json:"runtime"`

	// NumChars is the number of online characters that are currently
	// being tracked.
	NumChars int `json:"numchars"`

	// Err is holds any errors encountered by the monitor.
	Err error `json:"err,omitempty"`
}

// LoadSession loads a session from the file at path. It returns the
// session and an error, if any.
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

// Save saves the session to a file at path. The file is created if it
// doesn't exist, and truncated if it does. It returns an error if any
// are encountered.
func (s Session) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	e := json.NewEncoder(file)
	return e.Encode(&s)
}

// timeDiff is a light wrapper around time.Time that marshals to JSON
// as a duration since the time that the diff represents. For example,
// if time.Now() is 3 seconds after the time represented by the
// timeDiff, the JSON representation will be "3s".
type timeDiff time.Time

// Since returns the duration representing the difference between the
// current time and t.
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
