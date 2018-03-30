package lastfm

// An Artist represents the person or group responsible for creating a track
// or album.
type Artist struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	MusicBrainzID string `json:"mbid"`
}
