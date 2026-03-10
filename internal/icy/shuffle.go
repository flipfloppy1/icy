package icy

import (
	"errors"
	"math/rand"
	"net/http"
	"path/filepath"
)

func Shuffle(glob string, details RadioDetails) (http.Handler, error) {
	songs, err := filepath.Glob(glob)
	if err != nil {
		return nil, errors.New("icy.Shuffle: invalid glob")
	}

	if len(songs) == 0 {
		return nil, errors.New("icy.Shuffle: glob returned no matches")
	}

	handler := &playHandler{
		songs:        songs,
		radioDetails: details,
		subscribers:  make(map[string]chan chunk),
	}
	handler.nextFunc = shuffle
	go handler.tick()
	return handler, nil
}

// shuffle implements nextFunc for a shuffle playHandler.
func shuffle(n int) []int {
	next := make([]int, n)
	for i := range next {
		next[i] = i
	}
	rand.Shuffle(n, func(i int, j int) {
		currI := next[i]
		next[i] = next[j]
		next[j] = currI
	})
	return next
}
