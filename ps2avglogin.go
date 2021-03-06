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
	db, err := createDB()
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	log.Println("Loading session...")
	s, err := db.LoadSession()
	if err != nil {
		log.Printf("Failed to load session: %v", err)
		log.Println("Creating new session...")
	}
	s.Runtime = timeDiff(time.Now())
	if s.ShortestLong == 0 {
		s.ShortestLong = jsonDuration(1000 * time.Hour)
	}
	if s.Shortest == 0 {
		s.Shortest = jsonDuration(1000 * time.Hour)
	}

	copySession := func() Session {
		s := s
		s.NumChars = db.NumChar()

		return s
	}

	var oldest int64

	for {
		select {
		case ev := <-logins:
			t := time.Unix(ev.Timestamp, 0)

			err := db.SetChar(ev.CharacterID, t)
			if err != nil {
				// Not a fatal error.
				log.Printf("Failed to add %v to DB: %v", ev.CharacterID, err)
			}

			if oldest == 0 {
				oldest = ev.CharacterID

				s.Oldest = timeDiff(t)
				s.OldestName, err = getName(ev.CharacterID)
				if err != nil {
					log.Printf("Failed to get name for %v: %v", ev.CharacterID, err)
				}
				log.Printf("New oldest session is %q (%v) since %v", s.OldestName, ev.CharacterID, time.Time(s.Oldest))
			}

		case ev := <-logouts:
			in, ok, err := db.GetChar(ev.CharacterID)
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

					s.LongestName, err = getName(ev.CharacterID)
					if err != nil {
						log.Printf("Failed to get name for %v: %v", ev.CharacterID, err)
					}
					log.Printf("New longest session record is held by %q (%v) at %v", s.LongestName, ev.CharacterID, time.Duration(s.Longest))
				}
				if d < time.Duration(s.Shortest) {
					s.Shortest = jsonDuration(d)
				}

				err := db.RemoveChar(ev.CharacterID)
				if err != nil {
					log.Printf("Failed to remove %v from DB: %v", ev.CharacterID, err)
				}

				if ev.CharacterID == oldest {
					id, t, err := db.OldestChar()
					if err != nil {
						log.Printf("Failed to get oldest char: %v", err)
					}

					log.Printf("Previous oldest session was %q (%v) and lasted %v", s.OldestName, oldest, s.Oldest.String())
					oldest = id
					s.Oldest = timeDiff(t)
					s.OldestName, err = getName(id)
					if err != nil {
						log.Printf("Failed to get name for %v: %v", id, err)
					}
					log.Printf("New oldest session is %q (%v) since %v", s.OldestName, id, time.Time(s.Oldest))
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

	cancel := make(chan struct{})
	go autosave(cancel)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	log.Printf("Caught signal %q", <-sig)

	cancel <- struct{}{}
	<-cancel
}
