package sources

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
)

type ratings struct {
	Votes int `json:"votes"`
}

type traktMatch struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Ratings ratings `json:"ratings"`
}

type byRating []traktMatch

func (a byRating) Len() int           { return len(a) }
func (a byRating) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRating) Less(i, j int) bool { return a[i].Ratings.Votes > a[j].Ratings.Votes }

var traktSearchURL = "http://api.trakt.tv/search/shows.json/5bc6254d3bbde304a49557cf2845d921"

func constructUrl(query string) string {
	escapedQuery := url.Values{}
	escapedQuery.Add("query", query)
	return traktSearchURL + "?query=" + escapedQuery.Encode()
}

func Search(query string) ([]Match, error) {
	resp, err := http.Get(constructUrl(query))
	if err != nil {
		return nil, err //TODO retry a couple of times when it's a timeout.
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("Search return non 200 status code")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ms []traktMatch
	err = json.Unmarshal(body, &ms)
	if err != nil {
		return nil, err
	}

	sort.Sort(byRating(ms))

	return convertToMatches(ms), nil
}

func convertToMatches(ms []traktMatch) []Match {
	matches := make([]Match, len(ms))
	for i, m := range ms {
		matches[i] = Match{
			Title: m.Title,
			URL:   m.URL,
		}
	}
	return matches
}
