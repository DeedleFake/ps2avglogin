package main

import (
	"github.com/DeedleFake/census/ps2/events"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	SessionFile = "session.data"
)

var (
	logins  = make(chan *events.PlayerLogin)
	logouts = make(chan *events.PlayerLogout)

	average = make(chan time.Duration)
	session = make(chan *Session)
)

func coord() {
	log.Printf("Loading session from %q...", SessionFile)
	s, err := LoadSession(SessionFile)
	if err != nil {
		log.Printf("Failed to load session from %q: %v", SessionFile, err)
		log.Println("Creating new session...")
	}

	chars := make(map[int64]time.Time)

	for {
		select {
		case ev := <-logins:
			chars[ev.CharacterID] = time.Unix(ev.Timestamp, 0)
		case ev := <-logouts:
			if in, ok := chars[ev.CharacterID]; ok {
				s.Avg.Update(time.Unix(ev.Timestamp, 0).Sub(in))
				delete(chars, ev.CharacterID)
			}

		case average <- s.Avg.Cur:
		case session <- &s:
		}
	}
}

func monitor() {
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
			log.Fatalf("Error while fetching event: %v", err)
		}

		switch ev := ev.(type) {
		case *events.PlayerLogin:
			logins <- ev
		case *events.PlayerLogout:
			logouts <- ev
		}
	}
}

func main() {
	go monitor()
	go coord()
	go server()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Printf("Saving session to %q...", SessionFile)
	err := (<-session).Save(SessionFile)
	if err != nil {
		log.Printf("Failed to save session to %q: %v", SessionFile, err)
	}
}
