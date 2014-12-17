package sources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func Search(query string) Matches {
	resp, err := http.Get("http://api.trakt.tv/search/shows.json/5bc6254d3bbde304a49557cf2845d921?query=" + query)
	if err != nil { //TODO also error out on anything but a 200 Response
		fmt.Println("Error when searching: ", err)
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
