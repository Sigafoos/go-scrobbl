package lastfm

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	Version = 0.5
	BaseURL = "https://ws.audioscrobbler.com/2.0/"
)

var (
	SortOrder = []string{"0", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "1", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "2", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "3", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "4", "5", "6", "7", "8", "9"}
)

type LastFM struct {
	key        string
	secret     string
	sessionkey string
	client     *http.Client
	verbose    bool
}

func New(key, secret string, verbose bool) *LastFM {
	return &LastFM{
		key:     key,
		secret:  secret,
		verbose: verbose,
		client:  &http.Client{},
	}
}

func (l *LastFM) SessionKey(key string) {
	l.sessionkey = key
}

func (l *LastFM) Authenticated() bool {
	return l.sessionkey != ""
}

func (l *LastFM) Authenticate(username, password string) (string, error) {
	if l.Authenticated() {
		return "", fmt.Errorf("already authenticated")
	}

	query := fmt.Sprintf("api_key=%s&method=auth.getMobileSession&password=%s&username=%s", url.QueryEscape(l.key), url.QueryEscape(password), url.QueryEscape(username))
	query += fmt.Sprintf("&api_sig=%s&format=json", l.generateSignature(query))
	req, err := newHttpRequest("POST", BaseURL, bytes.NewBufferString(query))
	if err != nil {
		return "", err
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var response authenticationResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}
	return response.Session.Key, nil
}

func (l *LastFM) AlbumSearch(album string, requireMBID bool) ([]Album, error) {
	URL := fmt.Sprintf("%s?method=album.search&album=%s&api_key=%s&format=json", BaseURL, url.QueryEscape(album), l.key)
	if l.verbose {
		fmt.Printf("curl %s\n", URL)
	}
	req, err := newHttpRequest("GET", URL, nil)

	if err != nil {
		return []Album{}, err
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return []Album{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []Album{}, err
	}

	var searchResponse albumSearchResponse
	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		return []Album{}, err
	}

	if !requireMBID {
		return searchResponse.Results.AlbumMatches.Albums, nil
	}

	albums := []Album{}
	for _, v := range searchResponse.Results.AlbumMatches.Albums {
		if v.MusicBrainzID != "" {
			albums = append(albums, v)
		}
	}
	return albums, nil
}

func (l *LastFM) AlbumInfo(album Album) (Album, error) {
	var query string
	if album.MusicBrainzID != "" {
		query = fmt.Sprintf("mbid=%s", album.MusicBrainzID)
	} else {
		query = fmt.Sprintf("artist=%s&album=%s", url.QueryEscape(album.Artist), url.QueryEscape(album.Name))
	}

	URL := fmt.Sprintf("%s?%s&method=album.getInfo&api_key=%s&format=json", BaseURL, query, l.key)
	if l.verbose {
		fmt.Println(URL)
	}
	req, err := newHttpRequest("GET", URL, nil)

	if err != nil {
		return Album{}, err
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return Album{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Album{}, err
	}

	var searchResponse albumInfoResponse
	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		return Album{}, err
	}

	return searchResponse.Album, nil
}

func (l *LastFM) ScrobbleAlbum(album Album) error {
	if !l.Authenticated() {
		return fmt.Errorf("not authenticated")
	}

	// in order to generate the api signature we have to sort the keys in
	// alphabetical order, where the keys are string sorted
	mapped := make(map[string]Track)
	for _, v := range album.TrackList.Tracks {
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
			albums += fmt.Sprintf("&album[%s]=%v", k, url.QueryEscape(album.Name))
			tracknumbers += fmt.Sprintf("&tracknumber[%s]=%v", k, track.AttrList.TrackNumber)
			durations += fmt.Sprintf("&duration[%s]=%v", k, track.Duration)
		}
	}

	query := fmt.Sprintf("%s&api_key=%s%s%s&method=track.scrobble&sk=%s%s%s%s", albums, l.key, artists, durations, l.sessionkey, timestamps, tracks, tracknumbers)
	//query := fmt.Sprintf("api_key=%s%s&method=track.scrobble&sk=%s%s%s", l.key, artists, l.sessionkey, timestamps, tracks)
	unescaped, _ := url.QueryUnescape(query)
	query = fmt.Sprintf("api_sig=%s&%s&format=json", l.generateSignature(unescaped), query)
	req, err := newHttpRequest("POST", BaseURL, bytes.NewBufferString(query))
	if l.verbose {
		fmt.Printf("curl -X POST -d '%s' %s\n", query, BaseURL)
	}

	if err != nil {
		return err
	}

	resp, err := l.client.Do(req)
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

func newHttpRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return req, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("last.fm_go_library/%s", Version))

	return req, err
}

func (l *LastFM) generateSignature(query string) string {
	// this only works if your query is in alphabetical order
	r := strings.NewReplacer("?", "", "&", "", "=", "")
	query = r.Replace(query)
	query += l.secret
	return fmt.Sprintf("%x", md5.Sum([]byte(query)))
}
