package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
)

func getQuery() string {
	if len(os.Args) != 2 {
		fmt.Println("Please pass a search query.")
		os.Exit(1)
	}

	query := os.Args[1]
	return query
}

//TODO move this to a (search) providers package. Make sure the return type implements a certain interface.
type ratings struct {
	Votes int `json:"votes"`
}

type match struct {
	Title   string  `json:"title"`
	Ratings ratings `json:"ratings"`
}

type byRating []match

func (a byRating) Len() int           { return len(a) }
func (a byRating) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRating) Less(i, j int) bool { return a[i].Ratings.Votes > a[j].Ratings.Votes }

func search(query string) []match {
	resp, err := http.Get("http://api.trakt.tv/search/shows.json/5bc6254d3bbde304a49557cf2845d921?query=" + query)
	if err != nil { //TODO also error out on anything but a 200 Response
		fmt.Println("Error when searching: ", err)
		os.Exit(1) //TODO retry a couple of times when it's a timeout.
	}
	defer resp.Body.Close()

	fmt.Printf("resp %+v\n", resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error when parsing response: ", err)
		os.Exit(1)
	}

	var matches []match
	err = json.Unmarshal(body, &matches)
	if err != nil {
		fmt.Println("Error unmarshaling response: ", err)
	}

	sort.Sort(byRating(matches))
	return matches
}

func main() {
	query := getQuery()
	matches := search(query)
	fmt.Printf("matches %+v\n", matches)
}
