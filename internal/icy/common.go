package icy

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type audioInfo struct {
	Channels int
	SR       int
	BR       int
}

func (a audioInfo) String() string {
	return fmt.Sprintf("ice-channels=%d;ice-samplerate=%d;ice-bitrate=%d", a.Channels, a.SR, a.BR)
}

type RadioDetails struct {
	Name        string
	Description string
	Genre       string
	URL         string
	Public      bool
	metaint     int
}

type playHandler struct {
	songs            []string
	radioDetails     RadioDetails
	next             []int
	nextFunc         func(numSongs int) (order []int)
	currentIdx       int
	subscribers      map[string]chan chunk
	subscribersMutex sync.Mutex
}

type chunk struct {
	Data       []byte
	TagData    TagData
	SampleRate int
	BitRate    int
	Channels   int
}

// insertStreamTitle inserts icecast in-band metadata into data every h.metaint bytes, starting at offset.
//
// It returns the modified data and the new offset.
func (h *playHandler) insertStreamTitle(data []byte, title string, artist string, offset int) ([]byte, int) {
	if len(title) > 240 {
		title = title[:237] + "..."
	}

	metadata := []byte("StreamTitle='" + title + "';")
	if artist != "" {
		metadata = []byte("StreamTitle='" + artist + " - " + title + "';")
	}
	sizeByte := uint8(math.Ceil(float64(len(metadata)) / 16))
	paddingLen := int(sizeByte)*16 - len(metadata)
	metadata = append([]byte{sizeByte}, append(metadata, make([]byte, paddingLen)...)...)

	idx := offset

	for idx < len(data) {
		data = append(data[:idx], append(metadata, data[idx:]...)...)
		idx += h.radioDetails.metaint + len(metadata)
	}

	return data, idx - len(data)
}

// insertBlankMetadata inserts the icecast in-band metadata byte 0x00 every h.metaint bytes, starting at offset.
//
// It returns the modified data and the new offset.
func (h *playHandler) insertBlankMetadata(data []byte, offset int) ([]byte, int) {
	sizeByte := byte(0x00)
	idx := offset

	for idx < len(data) {
		data = append(data[:idx], append([]byte{sizeByte}, data[idx:]...)...)
		idx += h.radioDetails.metaint + 1
	}

	return data, idx - len(data)
}

func (h *playHandler) tick() {
	if h.radioDetails.metaint == 0 {
		h.radioDetails.metaint = 8192
	}

	h.next = h.nextFunc(len(h.songs))
	h.nextTrack()

	var tagData TagData
	for {
		f, err := os.Open(h.songs[h.currentIdx])
		if err != nil {
			fmt.Printf("icy: error opening MP3 file %s: %v\n", h.songs[h.currentIdx], err)
			continue
		}
		tagData = TitleHandler(f)
		f.Close()

		ffmpeg := exec.Command("ffmpeg", "-re", "-i", h.songs[h.currentIdx], "-map", "0:a", "-ar", "44100", "-ac", "2", "-b:a", "192k", "-f", "mp3", "pipe:1")

		stdout, err := ffmpeg.StdoutPipe()
		if err != nil {
			fmt.Printf("icy: error connecting stdout to ffmpeg: %v\n", err)
			continue
		}

		err = ffmpeg.Start()
		if err != nil {
			fmt.Printf("icy: error starting ffmpeg: %v\n", err)
			continue
		}

		go func() {
			for err == nil {
				data := make([]byte, 627)
				var n int
				n, err = stdout.Read(data)
				if n > 0 {
					h.sendToSubscribers(chunk{
						BitRate:    192,
						SampleRate: 44100,
						TagData:    tagData,
						Channels:   2,
						Data:       data[:n],
					})
				}
			}
		}()

		err = ffmpeg.Wait()
		if err != nil {
			fmt.Printf("icy: error waiting for ffmpeg to complete: %v\n", err)
		}

		if len(h.next) == 0 {
			h.next = h.nextFunc(len(h.songs))
		}

		h.nextTrack()
	}
}

// sendToSubscribers sends a clone of c to all subscribers.
func (h *playHandler) sendToSubscribers(c chunk) {
	h.subscribersMutex.Lock()
	for _, sub := range h.subscribers {
		select {
		case sub <- chunk{
			BitRate:    c.BitRate,
			TagData:    c.TagData,
			SampleRate: c.SampleRate,
			Channels:   c.Channels,
			Data:       bytes.Clone(c.Data), // cloned so it can be mutated by sub
		}:
		default:
		}
	}
	h.subscribersMutex.Unlock()
}

func (h *playHandler) nextTrack() {
	h.currentIdx = h.next[0]
	h.next = h.next[1:]
}

func (h *playHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "streaming not supported", http.StatusInternalServerError)
		return
	}
	subscriberId := rand.Text() + rand.Text()
	sub := make(chan chunk, 32)

	go func() {
		ctx := req.Context()
		<-ctx.Done()
		h.subscribersMutex.Lock()
		close(h.subscribers[subscriberId])
		delete(h.subscribers, subscriberId)
		h.subscribersMutex.Unlock()
	}()

	h.subscribersMutex.Lock()
	h.subscribers[subscriberId] = sub
	h.subscribersMutex.Unlock()

	first := <-sub

	if req.Header.Get("Icy-Metadata") != "1" {
		res.Header().Add("Content-Type", "text/plain")
		res.WriteHeader(400)
		res.Write([]byte("Expected an Icy-Metadata header with content '1'"))
		return
	}

	res.Header().Add("Content-Type", "audio/mpeg")
	res.Header().Add("Transfer-Encoding", "chunked")
	res.Header().Add("Connection", "keep-alive")
	res.Header().Add("icy-name", h.radioDetails.Name)
	res.Header().Add("icy-description", h.radioDetails.Description)
	res.Header().Add("icy-url", h.radioDetails.URL)
	res.Header().Add("icy-genre", h.radioDetails.Genre)
	pub := "0"
	if h.radioDetails.Public {
		pub = "1"
	}
	res.Header().Add("icy-pub", pub)
	res.Header().Add("icy-sr", strconv.Itoa(first.SampleRate))
	res.Header().Add("icy-br", strconv.Itoa(first.BitRate))
	res.Header().Add("icy-metaint", strconv.Itoa(h.radioDetails.metaint))
	res.Header().Add("icy-audio-info", audioInfo{Channels: first.Channels, BR: first.BitRate, SR: first.SampleRate}.String())

	var data []byte
	var offset int = h.radioDetails.metaint

	data, offset = h.insertStreamTitle(first.Data, first.TagData.TrackName, first.TagData.ArtistName, offset)

	_, err := res.Write(data)
	if err != nil {
		fmt.Printf("icy: error writing first chunk: %v\n", err)
		return
	}

	prevTitle := first.TagData.TrackName
	if len(first.Data) == len(data) {
		prevTitle = ""
	}

	for c := range sub {
		var data []byte
		if c.TagData.TrackName == prevTitle {
			data, offset = h.insertBlankMetadata(c.Data, offset)
		} else {
			data, offset = h.insertStreamTitle(c.Data, c.TagData.TrackName, c.TagData.ArtistName, offset)
			if len(c.Data) != len(data) {
				prevTitle = c.TagData.TrackName
			}
		}

		_, err := res.Write(data)
		if err != nil {
			return
		}
		flusher.Flush()
	}
}
