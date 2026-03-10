# icy

A dead simple Icecast transmission library for Go,
with support for in-band metadata.
Icy will figure out the song's title from id3v2 tags in MP3s, or use the
filename as a fallback. This behaviour can be changed with `SetTitleHandler`.

Icy only supports MP3, and expects all MP3 files to use the same sample rate and
bit rate (the latter is a limitation of Icecast).
It seems like some clients ignore the header-provided bitrate and sample rate
however, and use the MP3 frame bit/sample rates.

## API

See the [pkgsite documentation](pkg.go.dev/github.com/flipfloppy1/icy).

## Examples

```go
// shuffles all MP3 files in every subdirectory of mysongs
icy.Shuffle("./mysongs/*/*.mp3", icy.RadioDetails{})

// plays all MP3 files in the current directory in order
icy.Play("*.mp3", icy.RadioDetails{})
```
