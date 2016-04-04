package main

import (
	"bufio"
	"github.com/DeedleFake/census/ps2/events"
	"log"
	"os"
	"time"
)

var (
	logins  = make(chan *events.PlayerLogin)
	logouts = make(chan *events.PlayerLogout)

	average = make(chan time.Duration)
)

func coord() {
	var avg RollingAverage
	chars := make(map[int64]time.Time)

	for {
		select {
		case ev := <-logins:
			chars[ev.CharacterID] = time.Unix(ev.Timestamp, 0)
		case ev := <-logouts:
			if in, ok := chars[ev.CharacterID]; ok {
				avg.Update(time.Unix(ev.Timestamp, 0).Sub(in))
				delete(chars, ev.CharacterID)
			}

		case average <- avg.Get():
		}
	}
}

func monitor() {
	cl, err := events.NewClient("", "", "example")
	if err != nil {
		log.Fatalf("Failed to open client: %v", err)
	}
	defer cl.Close()

	cl.Subscribe(events.Sub{
		Events: []string{"PlayerLogin", "PlayerLogout"},
		Chars:  events.SubAll,
		Worlds: events.SubAll,
	})

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

	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		log.Printf("Average time spent online: %v", <-average)
	}
}
