package main

import (
	"bytes"
	"github.com/DeedleFake/census"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	client = &census.Client{
		Game: "ps2",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
)

func getName(out chan<- string, id int64, done chan<- struct{}) {
	defer close(done)

	var data struct {
		Chars []struct {
			Name struct {
				First string
			}
			Outfit struct {
				Alias string
			}
		} `json:"character_list"`
	}
	err := client.Get(&data,
		"character",
		census.SearchOption("character_id", strconv.FormatInt(id, 10)),
		census.ResolveOption("outfit"),
	)
	if err != nil {
		log.Printf("Failed to get character name for %v: %v", id, err)
		out <- "Error: Could not get name."
		return
	}
	if len(data.Chars) == 0 {
		out <- "Error: Could not get name."
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(data.Chars[0].Name.First)+len(data.Chars[0].Outfit.Alias)+3))
	if data.Chars[0].Outfit.Alias != "" {
		buf.WriteByte('[')
		buf.WriteString(data.Chars[0].Outfit.Alias)
		buf.WriteString("] ")
	}
	buf.WriteString(data.Chars[0].Name.First)

	out <- buf.String()
}
