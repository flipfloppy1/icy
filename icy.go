/*
Package icy is a dead simple Icecast transmission library for Go,
with support for in-band metadata.
*/
package icy

import (
	"net/http"
	"os"

	"github.com/flipfloppy1/icy/internal/icy"
)

type RadioDetails = icy.RadioDetails

// DefaultTitleHandler is the default function called to determine a song's name from its content/filename.
var DefaultTitleHandler = icy.DefaultTitleHandler

// Shuffle returns an [http.Handler] that streams all files collected by glob indefinitely in a randomized order.
func Shuffle(glob string, radioDetails RadioDetails) (http.Handler, error) {
	return icy.Shuffle(glob, radioDetails)
}

// Play returns an [http.Handler] that streams all files collected by glob indefinitely in order.
func Play(glob string, radioDetails RadioDetails) (http.Handler, error) {
	return icy.Play(glob, radioDetails)
}

// SetTitleHandler changes the function that is called to determine a song's name from its content/filename.
func SetTitleHandler(titleHandler func(*os.File) icy.TagData) {
	icy.TitleHandler = titleHandler
}
