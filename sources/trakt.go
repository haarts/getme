// TODO this should just be a Trakt client. Minorring the API endpoints from Trakt.
package sources

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/haarts/getme/store"
)

// Trakt is the struct which implements the Source interface.
type Trakt struct{}

// TRAKT defines the name of this source.
const traktName = "trakt"

func init() {
	Register(traktName, Trakt{})
}

type traktSeason struct {
	Season   int `json:"number"`
	Episodes int `json:"episode_count"`
}

type traktEpisode struct {
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	FirstAired *time.Time `json:"first_aired"`
}

type traktShow struct {
	// WHat to do what to do?
}

type traktSearchResult struct {
	Score float64 `json:"score"`
	Show  struct {
		Title string `json:"title"`
		IDs   struct {
			Slug string `json:"slug"`
		}
	} `json:"show"`
}

type byScore []traktSearchResult

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].Score > a[j].Score }

// NOTE Yes, you can use this access token for you're own app...
var accessToken = "3b6f5bdba2fa56b086712d5f3f15b4e967f99ab049a6d3a4c2e56dc9c3c90462"
var clientID = "01045164ed603042b53acf841b590f0e7b728dbff319c8d128f8649e2427cbe9" //AKA trakt-api-key
var traktURL = "https://api.trakt.tv"
var traktSearchURL = traktURL + "/search?type=show"

func (t Trakt) seasonURL(show store.Show, season store.Season) string {
	return fmt.Sprintf(traktURL+"/shows/%s/seasons/%d?extended=full", show.URL, season.Season)
}

func (t Trakt) seasonsURL(show store.Show) string {
	return fmt.Sprintf(traktURL+"/shows/%s/seasons?extended=full", show.URL)
}

func searchShowURL(query string) string {
	escapedQuery := url.Values{}
	escapedQuery.Add("query", query)
	return traktSearchURL + "&" + escapedQuery.Encode()
}

func traktRequest(URL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("trakt-api-key", clientID)

	return req, nil
}

// Search returns matches found by this source based on the query.
func (t Trakt) Search(query string) ([]Match, error) {
	req, err := traktRequest(searchShowURL(query))
	if err != nil {
		return nil, err
	}

	ms := &([]traktSearchResult{})

	err = GetJSON(req, ms)
	if err != nil {
		return nil, err
	}

	sort.Sort(byScore(*ms))

	return convertToMatches(*ms), nil
}

// TODO expand Matches construct to incorporate a Show. Hell, let's not return
// a Match here but a TraktShow.
func convertToMatches(ms []traktSearchResult) []Match {
	matches := make([]Match, len(ms))
	for i, m := range ms {
		matches[i] = store.Show{
			URL:        m.Show.IDs.Slug,
			Title:      m.Show.Title,
			SourceName: traktName,
		}
	}
	return matches
}

// AllSeasonsAndEpisodes finds the seasons and episodes for a show with this source.
// TODO get rid of anything store related. Return own types and let the caller
// decide what to do with it.
func (t Trakt) AllSeasonsAndEpisodes(show store.Show) ([]*store.Season, error) {
	req, err := traktRequest(t.seasonsURL(show))
	if err != nil {
		return nil, err
	}

	ss := &([]traktSeason{})

	err = GetJSON(req, ss)
	if err != nil {
		return nil, err
	}

	seasons := convertToSeasons(*ss)
	err = t.addEpisodes(seasons, show)
	if err != nil {
		return nil, err
	}
	return seasons, nil
}

// TODO Quite a bit of duplication with the convertToMatches function.
func convertToSeasons(ss []traktSeason) []*store.Season {
	seasons := make([]*store.Season, 0, len(ss))
	for _, s := range ss {
		season := &store.Season{
			Season:   s.Season,
			Episodes: make([]*store.Episode, 0, s.Episodes),
		}
		seasons = append(seasons, season)
	}
	return seasons
}

func (t Trakt) addEpisodes(seasons []*store.Season, show store.Show) error {
	for _, season := range seasons {
		req, err := traktRequest(t.seasonURL(show, *season))
		if err != nil {
			return err
		}

		episodes := &([]traktEpisode{})
		err = GetJSON(req, episodes)
		if err != nil {
			return err
		}

		for _, episode := range *episodes {
			if episode.FirstAired == nil {
				episode.FirstAired = &time.Time{}
			}
			e := store.Episode{
				Title:   episode.Title,
				Episode: episode.Number,
				Pending: true, // NOTE Do not forget to set pending to true!
				AirDate: *episode.FirstAired,
			}

			season.Episodes = append(season.Episodes, &e)
		}
	}
	return nil
}
