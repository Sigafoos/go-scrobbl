package lastfm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"time"
)

// An Album represents an album, and implements the Scrobbler interface.
type Album struct {
	API           *API
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

// A TrackList contains a slice of Tracks.
type TrackList struct {
	Tracks []Track `json:"track"`
}

// An Image represents the cover art of an album.
type Image struct {
	URL  string `json:"#text"`
	Size string `json:"small"`
}

// GetInfo uses the information in the Album to retrieve more information
// (including the track list) from last.fm. It uses the MusicBrainz ID, if
// available, as that is more authoritative.
func (a *Album) GetInfo() error {
	if a.API == nil {
		return fmt.Errorf("no API object associated with Album")
	}

	var query string
	if a.MusicBrainzID != "" {
		query = fmt.Sprintf("mbid=%s", a.MusicBrainzID)
	} else {
		if a.Artist == "" || a.Name == "" {
			return fmt.Errorf("GetInfo requires either a MBID or artist and album")
		}
		query = fmt.Sprintf("artist=%s&album=%s", url.QueryEscape(a.Artist), url.QueryEscape(a.Name))
	}

	URL := fmt.Sprintf("%s?%s&method=album.getInfo&api_key=%s&format=json", BaseURL, query, a.API.key)
	if a.API.verbose {
		fmt.Println(URL)
	}
	req, err := a.API.newHTTPRequest("GET", URL, nil)

	if err != nil {
		return err
	}

	resp, err := a.API.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var searchResponse albumInfoResponse
	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		return err
	}

	// overwrite everything entered with the canonical response
	a.Name = searchResponse.Album.Name
	a.Artist = searchResponse.Album.Artist
	a.URL = searchResponse.Album.URL
	a.Images = searchResponse.Album.Images
	a.Streamable = searchResponse.Album.Streamable
	a.MusicBrainzID = searchResponse.Album.MusicBrainzID
	a.Listeners = searchResponse.Album.Listeners
	a.Playcount = searchResponse.Album.Playcount
	a.TrackList = searchResponse.Album.TrackList

	return nil
}

// Scrobble send a request to last.fm to scrobble all tracks on the album. It
// sends one API call using the batch format.
func (a *Album) Scrobble() error {
	if !a.API.Authenticated() {
		return fmt.Errorf("not authenticated")
	}

	// in order to generate the api signature we have to sort the keys in
	// alphabetical order, where the keys are string sorted
	mapped := make(map[string]Track)
	for _, v := range a.TrackList.Tracks {
		s := strconv.Itoa(v.AttrList.TrackNumber - 1)
		mapped[s] = v
	}

	start := time.Now().Unix()
	var artists, tracks, timestamps, albums, tracknumbers, durations string
	for _, k := range SortOrder {
		if track, ok := mapped[k]; ok {
			artists += fmt.Sprintf("&artist[%s]=%s", k, track.Artist.Name)
			tracks += fmt.Sprintf("&track[%s]=%s", k, url.QueryEscape(track.Name))
			timestamps += fmt.Sprintf("&timestamp[%s]=%v", k, start+int64(track.AttrList.TrackNumber))
			albums += fmt.Sprintf("&album[%s]=%v", k, url.QueryEscape(a.Name))
			tracknumbers += fmt.Sprintf("&tracknumber[%s]=%v", k, track.AttrList.TrackNumber)
			durations += fmt.Sprintf("&duration[%s]=%v", k, track.Duration)
		}
	}

	query := fmt.Sprintf("%s&api_key=%s%s%s&method=track.scrobble&sk=%s%s%s%s", albums, a.API.key, artists, durations, a.API.sessionkey, timestamps, tracks, tracknumbers)
	unescaped, _ := url.QueryUnescape(query)
	query = fmt.Sprintf("api_sig=%s&%s&format=json", a.API.generateSignature(unescaped), query)
	req, err := a.API.newHTTPRequest("POST", BaseURL, bytes.NewBufferString(query))
	if err != nil {
		return err
	}

	if a.API.verbose {
		fmt.Printf("curl -X POST -d '%s' %s\n", query, BaseURL)
	}

	resp, err := a.API.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var scrobbleResp scrobbleResponse
	err = json.Unmarshal(body, &scrobbleResp)
	if err != nil {
		return err
	} else if scrobbleResp.Error != 0 {
		return fmt.Errorf("%v", scrobbleResp.Error)
	}

	return nil
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
