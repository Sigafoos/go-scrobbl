// Package lastfm provides a way to communicate with last.fm's api in go.
package lastfm

var (
	// Version is the SemVer representation of the module's state.
	Version = "0.6.1"
	// BaseURL is the root of all API calls.
	BaseURL = "https://ws.audioscrobbler.com/2.0/"
)

// SortOrder is the order array keys should be in to generate a valid signature. The "signature" for authenticated calls requires you to put the parameters in
// alphabetical order. This includes the arrays in track.scrobble, which have to be sorted according to
// the ASCII table, not as though the numbers were strings. By this method, 10 < 19 < 1. Much time was
// spent trying to make `sort.Sort` work with this until I decided this was perfectly fine.
var SortOrder = []string{"0", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "1", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "2", "30", "31", "32", "33", "34", "35", "36", "37", "38", "39", "3", "40", "41", "42", "43", "44", "45", "46", "47", "48", "49", "4", "5", "6", "7", "8", "9"}

// A Scrobbler is a type that can scrobble itself, or record the track(s) it
// contains to last.fm.
type Scrobbler interface {
	Scrobble() error
}

// reeeeally need to add more
type scrobbleResponse struct {
	Error   int    `json:"error,omitempty"`
	Message string `json:"message"`
}
