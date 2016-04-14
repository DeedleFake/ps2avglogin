package main

import (
	"bytes"
	"github.com/DeedleFake/census"
	"net/http"
	"strconv"
	"time"
)

var (
	client = &census.Client{
		Game: "ps2",
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
)

func getName(id int64) (string, error) {
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
		return "", err
	}
	if len(data.Chars) == 0 {
		return "", noSuchCharError(id)
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(data.Chars[0].Name.First)+len(data.Chars[0].Outfit.Alias)+3))
	if data.Chars[0].Outfit.Alias != "" {
		buf.WriteByte('[')
		buf.WriteString(data.Chars[0].Outfit.Alias)
		buf.WriteString("] ")
	}
	buf.WriteString(data.Chars[0].Name.First)

	return buf.String(), nil
}

type noSuchCharError int64

func (err noSuchCharError) Error() string {
	return "No such char: " + strconv.FormatInt(int64(err), 10)
}
