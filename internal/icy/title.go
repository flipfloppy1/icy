package icy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/tmthrgd/id3v2"
)

type TagData struct {
	TrackName  string
	ArtistName string
}

var TitleHandler func(f *os.File) TagData = DefaultTitleHandler

func DefaultTitleHandler(f *os.File) TagData {
	data := TagData{TrackName: filepath.Base(f.Name()), ArtistName: ""}

	frames, err := id3v2.Scan(f)
	if err != nil {
		return data
	}

	mapPrintable := func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}

	titleFrame := frames.Lookup(id3v2.FrameTIT2)
	artistFrame := frames.Lookup(id3v2.FrameTPE1)

	if titleFrame != nil {
		data.TrackName = strings.Map(mapPrintable, string(titleFrame.Data))
	} else {
		fmt.Println("id3v2 tags did not contain TIT1")
	}

	if artistFrame != nil {
		data.ArtistName = strings.Map(mapPrintable, string(artistFrame.Data))
	}

	return data
}
