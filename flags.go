package main

import (
	"bytes"
	"flag"
	"fmt"
	"strings"
	"time"
)

// flags stores the command line flags and arguments.
var flags struct {
	addr    string
	session string
	short   time.Duration
	db      mapFlag
}

func init() {
	flags.short = time.Hour
	flags.db = mapFlag{"type": "map"}

	flag.StringVar(&flags.addr, "addr", ":8080", "The address to run the web interface at.")
	flag.StringVar(&flags.session, "s", "session.json", "The session file to use.")
	flag.Var((*durationFlag)(&flags.short), "short", "The maximum length of a session to consider short.")
	flag.Var(&flags.db, "db", "Options for the database.")

	flag.Parse()
}

// durationFlag is a wrapper around time.Duration to make it satisfy
// flag.Value.
type durationFlag time.Duration

func (f durationFlag) String() string {
	return time.Duration(f).String()
}

func (f *durationFlag) Set(val string) error {
	d, err := time.ParseDuration(val)
	*f = durationFlag(d)
	return err
}

type mapFlag map[string]string

func (f mapFlag) String() string {
	var buf bytes.Buffer
	for k, v := range f {
		buf.WriteByte(',')
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(v)
	}
	buf.ReadByte()

	return buf.String()
}

func (f mapFlag) Set(val string) error {
	opts := strings.Split(val, ",")
	for _, opt := range opts {
		kv := strings.SplitN(opt, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("Invalid flag: %q", kv)
		}

		f[kv[0]] = kv[1]
	}

	return nil
}
