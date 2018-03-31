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
	"strings"
)

// An API object provides the mechanism to authenticate and communicate with last.fm's system.
type API struct {
	key        string
	secret     string
	sessionkey string
	client     *http.Client
	verbose    bool
}

// New initializes an API object with the provided api key and shared secret. The user's session key is not
// required as you may need to contact the API in order to obtain one.
func New(key, secret string) *API {
	return &API{
		key:    key,
		secret: secret,
		client: &http.Client{},
	}
}

// SetSessionKey adds the user's session key/token to the object, allowing for authenticated API calls.
func (a *API) SetSessionKey(key string) {
	a.sessionkey = key
}

// SetVerbose sets whether more information (such as curl versions of the calls).
func (a *API) SetVerbose(verbose bool) {
	a.verbose = verbose
}

// Authenticated returns whether a session key has been set.
func (a *API) Authenticated() bool {
	return a.sessionkey != ""
}

// Authenticate accepts a username and password and uses the mobile authentication
// scheme to generate a session token. Using the mobile scheme avoids the need
// to handle oauth callbacks.
func (a *API) Authenticate(username, password string) (string, error) {
	if a.Authenticated() {
		return "", fmt.Errorf("already authenticated")
	}

	query := fmt.Sprintf("api_key=%s&method=auth.getMobileSession&password=%s&username=%s", a.key, password, username)
	query += fmt.Sprintf("&api_sig=%s&format=json", a.generateSignature(query))
	req, err := a.newHTTPRequest("POST", BaseURL, bytes.NewBufferString(query))
	if err != nil {
		return "", err
	}

	resp, err := a.client.Do(req)
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

// AlbumSearch contacts last.fm and returns a list of albums that match the
// search query.
func (a *API) AlbumSearch(album string, requireMBID bool) ([]Album, error) {
	URL := fmt.Sprintf("%s?method=album.search&album=%s&api_key=%s&format=json", BaseURL, url.QueryEscape(album), a.key)
	if a.verbose {
		fmt.Printf("curl %s\n", URL)
	}
	req, err := a.newHTTPRequest("GET", URL, nil)

	if err != nil {
		return []Album{}, err
	}

	resp, err := a.client.Do(req)
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

	albums := []Album{}
	for _, v := range searchResponse.Results.AlbumMatches.Albums {
		if requireMBID && v.MusicBrainzID == "" {
			continue
		}
		v.API = a
		albums = append(albums, v)
	}
	return albums, nil
}

func (a *API) newHTTPRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return req, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("last.fm_go_library/%s", Version))

	return req, err
}

func (a *API) generateSignature(query string) string {
	// this only works if your query is in alphabetical order
	r := strings.NewReplacer("?", "", "&", "", "=", "")
	query = r.Replace(query)
	query += a.secret
	return fmt.Sprintf("%x", md5.Sum([]byte(query)))
}

type authenticationResponse struct {
	Session sessionResponse `json:"session"`
}

type sessionResponse struct {
	Subscriber int    `json:"subscriber"`
	Name       string `json:"name"`
	Key        string `json:"key"`
}
