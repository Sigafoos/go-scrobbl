[![Go Report Card](https://goreportcard.com/badge/github.com/Sigafoos/lastfm)](https://goreportcard.com/report/github.com/Sigafoos/lastfm)
[![GoDoc](https://godoc.org/github.com/Sigafoos/lastfm?status.svg)](https://godoc.org/github.com/Sigafoos/lastfm)
# lastfm
Go package to interface with last.fm

## Usage
Instantiate the API object with your API key and secret.

	lfm := lastfm.New("yourkey", "yoursecret")
	token, err := lfm.Authenticate("yourusername", "yourpassword")
	if err != nil {
		lfm.SetSessionKey(token)
	}

Search for an album (requiring a [MusicBrainz](https://musicbrainz.org/) ID) and scrobble the first result.

	albums, err := lfm.AlbumSearch("slow riot for new zero kanada", true)
	if err != nil {
		err = albums[0].Scrobble()
	}

Obviously it makes sense to not just blindly trust that the first result is what you want. See [https://github.com/Sigafoos/scrobble](github.com/Sigafoos/scrobble) for a CLI utilizing this package.
