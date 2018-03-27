package lastfm

import ()

type authenticationResponse struct {
	Session sessionResponse `json:"session"`
}

type sessionResponse struct {
	Subscriber int    `json:"subscriber"`
	Name       string `json:"name"`
	Key        string `json:"key"`
}

type albumSearchResponse struct {
	Results albumSearchResults `json:"results"`
}

type albumSearchResults struct {
	TotalResults   string       `json:"opensearch:totalResults"`
	Start          string       `json:"opensearch:startIndex"`
	ResultsPerPage string       `json:"opensearch:itemsPerPage"`
	AlbumMatches   albumMatches `json:"albummatches"`
}

type albumMatches struct {
	Albums []Album `json:"album"`
}

type albumInfoResponse struct {
	Album Album `json:"album"`
}

type Album struct {
	Name          string    `json:"name"`
	Artist        string    `json:"artist"`
	URL           string    `json:"url"`
	Images        []Image   `json:"image"`
	Streamable    string    `json:"streamable"`
	MusicBrainzID string    `json:"mbid"`
	Listeners     string    `json:"lisners"`
	Playcount     string    `json:"playcount"`
	TrackList     TrackList `json:"tracks"`
}

type Artist struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	MusicBrainzID string `json:"mbid"`
}

type TrackList struct {
	Tracks []Track `json:"track"`
}

type Track struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Duration int      `json:"duration,string"`
	AttrList AttrList `json:"@attr"`
	Artist   Artist   `json:"artist"`
	// not including streamable
}

type Image struct {
	URL  string `json:"#text"`
	Size string `json:"small"`
}

type AttrList struct {
	TrackNumber int `json:"rank,string"`
}

// reeeeally need to add more
type scrobbleResponse struct {
	Error int `json:"error,omitempty"`
}
