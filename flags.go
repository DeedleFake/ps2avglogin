package main

import (
	"flag"
	"time"
)

var flags struct {
	addr    string
	session string
	short   time.Duration
}

func init() {
	flags.short = time.Hour

	flag.StringVar(&flags.addr, "addr", ":8080", "The address to run the web interface at.")
	flag.StringVar(&flags.session, "s", "session.json", "The session file to use.")
	flag.Var((*durationFlag)(&flags.short), "short", "The maximum length of a session to consider short.")
	flag.Parse()
}

type durationFlag time.Duration

func (f durationFlag) String() string {
	return time.Duration(f).String()
}

func (f *durationFlag) Set(val string) error {
	d, err := time.ParseDuration(val)
	*f = durationFlag(d)
	return err
}
