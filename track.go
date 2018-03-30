package lastfm

import (
	"fmt"
)

// A Track represents a song and implements the Scrobbler interface.
type Track struct {
	API      *API
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Duration int      `json:"duration,string"`
	AttrList AttrList `json:"@attr"`
	Artist   Artist   `json:"artist"`
	// this doesn't include the streamable attribute
}

// An AttrList contains attributes about a track.
type AttrList struct {
	TrackNumber int `json:"rank,string"`
}

// Scrobble records the track on the user's last.fm account.
func (t *Track) Scrobble() error {
	return fmt.Errorf("Track.Scrobble() is not yet implemented")
}
