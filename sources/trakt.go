package sources

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
)

type Trakt struct{}

const TRAKT = "trakt"

func init() {
	Register(TRAKT, Trakt{})
}

type traktMatch struct {
	Score float64 `json:"score"`
	Show  struct {
		Title string `json:"title"`
		IDs   struct {
			Slug string `json:"slug"`
		}
	} `json:"show"`
}

type byScore []traktMatch

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].Score > a[j].Score }

var access_token = "3b6f5bdba2fa56b086712d5f3f15b4e967f99ab049a6d3a4c2e56dc9c3c90462"
var client_id = "01045164ed603042b53acf841b590f0e7b728dbff319c8d128f8649e2427cbe9" //AKA trakt-api-key
var traktURL = "https://api.trakt.tv"
var traktSearchURL = traktURL + "/search?type=show"

func constructURL(query string) string {
	escapedQuery := url.Values{}
	escapedQuery.Add("query", query)
	return traktSearchURL + "&" + escapedQuery.Encode()
}

func traktRequest(URL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+access_token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("trakt-api-key", client_id)

	return req, nil
}

func (t Trakt) Search(query string) ([]Match, error) {
	req, err := traktRequest(constructURL(query))
	fmt.Printf("req %+v\n", req)
	fmt.Printf("err %+v\n", err)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	fmt.Printf("err %+v\n", err)
	if err != nil {
		return nil, err //TODO retry a couple of times when it's a timeout.
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Search returned non 200 status code: %d", resp.StatusCode))
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

	sort.Sort(byScore(ms))

	return convertToMatches(ms), nil
}

func convertToMatches(ms []traktMatch) []Match {
	matches := make([]Match, len(ms))
	for i, m := range ms {
		matches[i] = Show{
			URL:        m.Show.IDs.Slug,
			Title:      m.Show.Title,
			SourceName: TRAKT,
		}
	}
	return matches
}
