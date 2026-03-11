# icy

A dead simple Icecast transmission library for Go, with support for in-band
metadata.
Icy will figure out the song's title from id3v2 tags in MP3s, or use the
filename as a fallback. This behaviour can be changed with `SetTitleHandler`.

Icy supports whatever input files ffmpeg can transcode. Incidentally, ffmpeg has
to be available on the system path for it to work.
The original solution decoded MP3 files using a pure Go library, but it didn't
work very well.

Icy uses ffmpeg to transcode all input files into a 44.1kHz/192kb double channel
MPEG stream.

## API

See the [pkgsite documentation](pkg.go.dev/github.com/flipfloppy1/icy).

## Examples

```go
// shuffles all MP3 files in every subdirectory of mysongs
icy.Shuffle("./mysongs/*/*.mp3", icy.RadioDetails{})

// plays all M4A files in the current directory in order
icy.Play("*.m4a", icy.RadioDetails{})
```
