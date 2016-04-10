package main

import (
	"github.com/DeedleFake/census/ps2/events"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	s.Longest = 0
	s.ShortestLong = jsonDuration(1000 * time.Hour)
	s.Shortest = jsonDuration(1000 * time.Hour) // Just need something ridiculous.

	chars, err := createDB()
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer chars.Close()

	copySession := func() Session {
		s := s
		s.NumChars = chars.Num()

		return s
	}

	for {
		select {
		case ev := <-logins:
			err := chars.Set(ev.CharacterID, time.Unix(ev.Timestamp, 0))
			if err != nil {
				// Not a fatal error.
				log.Printf("Failed to add %v to DB: %v", ev.CharacterID, err)
			}
		case ev := <-logouts:
			in, ok, err := chars.Get(ev.CharacterID)
			if err != nil {
				log.Printf("Failed to get %v from DB: %v", ev.CharacterID, err)
				continue
			}
			if ok {
				d := time.Unix(ev.Timestamp, 0).Sub(in)

				s.Total.Update(d)
				if d > flags.short {
					s.NoShort.Update(d)

					if d < time.Duration(s.ShortestLong) {
						s.ShortestLong = jsonDuration(d)
					}
				}

				if d > time.Duration(s.Longest) {
					s.Longest = jsonDuration(d)
				}
				if d < time.Duration(s.Shortest) {
					s.Shortest = jsonDuration(d)
				}

				err := chars.Remove(ev.CharacterID)
				if err != nil {
					log.Printf("Failed to remove %v from DB: %v", ev.CharacterID, err)
				}
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
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	log.Printf("Caught signal %q", <-sig)

	log.Printf("Saving session to %q...", flags.session)
	err := (<-session).Save(flags.session)
	if err != nil {
		log.Printf("Failed to save session to %q: %v", flags.session, err)
	}
}
