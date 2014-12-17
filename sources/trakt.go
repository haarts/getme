package sources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
)

type ratings struct {
	Votes int `json:"votes"`
}

type traktMatch struct {
	FoundTitle string  `json:"title"`
	Ratings    ratings `json:"ratings"`
}

func (t traktMatch) Title() string {
	return t.FoundTitle
}

type traktMatches []traktMatch

func (tm traktMatches) BestMatch() Match {
	return tm[0]
}

type byRating []traktMatch

func (a byRating) Len() int           { return len(a) }
func (a byRating) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRating) Less(i, j int) bool { return a[i].Ratings.Votes > a[j].Ratings.Votes }

var traktURL = "http://api.trakt.tv/search/shows.json/5bc6254d3bbde304a49557cf2845d921"

func Search(query string) Matches {
	escapedQuery := url.Values{}
	escapedQuery.Add("query", query)
	resp, err := http.Get(traktURL + "?query=" + escapedQuery.Encode())
	if err != nil || resp.StatusCode != 200 { //TODO Tidy this up (better logging)
		fmt.Println("Error when searching: ", err)
		fmt.Println("Error when searching: ", resp)
		os.Exit(1) //TODO retry a couple of times when it's a timeout.
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error when parsing response: ", err)
		os.Exit(1)
	}

	var ms []traktMatch
	err = json.Unmarshal(body, &ms)
	if err != nil {
		fmt.Println("Error unmarshaling response: ", err)
	}

	sort.Sort(byRating(ms))

	return traktMatches(ms)
}
