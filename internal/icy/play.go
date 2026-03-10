package icy

import (
	"errors"
	"net/http"
	"path/filepath"
)

func Play(glob string, details RadioDetails) (http.Handler, error) {
	songs, err := filepath.Glob(glob)
	if err != nil {
		return nil, errors.New("icy.Play: invalid glob")
	}

	if len(songs) == 0 {
		return nil, errors.New("icy.Play: glob returned no matches")
	}

	handler := &playHandler{
		songs:        songs,
		radioDetails: details,
		subscribers:  make(map[string]chan chunk),
	}
	handler.nextFunc = play
	go handler.tick()
	return handler, nil
}

// play implements nextFunc for an ordered playHandler.
func play(n int) []int {
	next := make([]int, n)
	for i := range next {
		next[i] = i
	}
	return next
}
