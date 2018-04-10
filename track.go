package lastfm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"
)

// A Track represents a song and implements the Scrobbler interface.
type Track struct {
	API      *API
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Duration int      `json:"duration,string"`
	AttrList AttrList `json:"@attr"`
	Artist   Artist   `json:"artist"`
	Album    string
	// this doesn't include the streamable attribute
}

// An AttrList contains attributes about a track.
type AttrList struct {
	TrackNumber int `json:"rank,string"`
}

// Scrobble records the track on the user's last.fm account.
func (t *Track) Scrobble() error {
	if !t.API.Authenticated() {
		return fmt.Errorf("not authenticated")
	} else if t.Duration > 0 && t.Duration < 30 {
		return fmt.Errorf("track cannot be less than 30 seconds (track duration is %v seconds)", t.Duration)
	}

	var query string
	if t.Album != "" {
		query += fmt.Sprintf("album[0]=%s&", url.QueryEscape(t.Album))
	}
	query += fmt.Sprintf("api_key=%s&artist[0]=%s", t.API.key, url.QueryEscape(t.Artist.Name))
	if t.Duration != 0 {
		query += fmt.Sprintf("&duration[0]=%v", t.Duration)
	}
	query += fmt.Sprintf("&method=track.scrobble&sk=%s&timestamp[0]=%v&track[0]=%s", t.API.sessionkey, time.Now().Unix(), url.QueryEscape(t.Name))
	unescaped, _ := url.QueryUnescape(query)
	query = fmt.Sprintf("api_sig=%s&%s&format=json", t.API.generateSignature(unescaped), query)
	req, err := t.API.newHTTPRequest("POST", BaseURL, bytes.NewBufferString(query))
	if err != nil {
		return err
	}

	if t.API.verbose {
		fmt.Printf("curl -X POST -d '%s' %s\n", query, BaseURL)
	}

	resp, err := t.API.client.Do(req)
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
