# icy

A dead simple Icecast transmission library for Go,
with support for in-band metadata.
Icy will pull metadata from MP3 if it exists, or use the filename if not.

It only supports MP3, and expects all MP3 files to use the same sample rate and
bit rate (the latter is a limitation of Icecast).

## API

See the [pkgsite documentation](pkg.go.dev/github.com/flipfloppy1/icy).

## Examples

```go
// shuffles all MP3 files in every subdirectory of mysongs
icy.Shuffle("./mysongs/*/*.mp3", icy.RadioDetails{})

// plays all wav files in the current directory in order
icy.Play("*.wav", icy.RadioDetails{})
```
