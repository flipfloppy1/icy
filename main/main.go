package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/flipfloppy1/icy"
)

func main() {
	handler, err := icy.Shuffle(
		"../../Sync/Music/abel cirilo/*/*.mp3",
		icy.RadioDetails{
			Name:        "Leechplus",
			Description: "Leechplus",
			Genre:       "Electronic",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "./index.html")
	})
	http.Handle("/radio", handler)

	err = http.ListenAndServe(":8080", nil)
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
