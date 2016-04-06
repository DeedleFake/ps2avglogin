package main

import (
	"github.com/DeedleFake/census/ps2/events"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	// session can be used to get a copy of the current session.
	session = make(chan Session)
)

// coord coordinates the session, updating it properly when login and
// logout events occur, and sending a copy of the session down the
// session channel when it's requested.
func coord(logins <-chan *events.PlayerLogin, logouts <-chan *events.PlayerLogout, errors <-chan error) {
	log.Printf("Loading session from %q...", flags.session)
	s, err := LoadSession(flags.session)
	if err != nil {
		log.Printf("Failed to load session from %q: %v", flags.session, err)
		log.Println("Creating new session...")
	}
	s.Runtime = timeDiff(time.Now())

	chars := make(map[int64]time.Time)

	copySession := func() Session {
		s := s
		s.NumChars = len(chars)

		return s
	}

	for {
		select {
		case ev := <-logins:
			chars[ev.CharacterID] = time.Unix(ev.Timestamp, 0)
		case ev := <-logouts:
			// TODO: Use the REST API to get login times, rather than
			// tracking it manually.
			if in, ok := chars[ev.CharacterID]; ok {
				d := time.Unix(ev.Timestamp, 0).Sub(in)

				s.Total.Update(d)
				if d > flags.short {
					s.NoShort.Update(d)
				}

				delete(chars, ev.CharacterID)
			}

		case err := <-errors:
			s.Err = err

		case session <- copySession():
		}
	}
}

// monitor connects to the census API, subscribes to PlayerLogin and
// PlayerLogout events, and then sends them down the appropriate
// channels.
func monitor(logins chan<- *events.PlayerLogin, logouts chan<- *events.PlayerLogout, errors chan<- error) {
	cl, err := events.NewClient("", "", "example")
	if err != nil {
		log.Fatalf("Failed to open client: %v", err)
	}
	defer cl.Close()

	err = cl.Subscribe(events.Sub{
		Events: []string{"PlayerLogin", "PlayerLogout"},
		Chars:  events.SubAll,
		Worlds: events.SubAll,
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to login/logout events: %v", err)
	}

	for {
		ev, err := cl.Next()
		if err != nil {
			log.Printf("Error while fetching event: %v", err)
			errors <- err
			continue
		}
		errors <- nil

		switch ev := ev.(type) {
		case *events.PlayerLogin:
			logins <- ev
		case *events.PlayerLogout:
			logouts <- ev
		}
	}
}

func main() {
	logins := make(chan *events.PlayerLogin)
	logouts := make(chan *events.PlayerLogout)
	errors := make(chan error)

	go monitor(logins, logouts, errors)
	go coord(logins, logouts, errors)
	go server()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Printf("Saving session to %q...", flags.session)
	err := (<-session).Save(flags.session)
	if err != nil {
		log.Printf("Failed to save session to %q: %v", flags.session, err)
	}
}
