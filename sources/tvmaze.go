package sources

import (
	"fmt"
	"net/http"

	"github.com/haarts/getme/store"
)

var tvMazeURL = "http://api.tvmaze.com"

const tvMazeName = "tvmaze"

type TvMaze struct{}

func (t TvMaze) Name() string {
	return tvMazeName
}

func (t TvMaze) Search(q string) SearchResult {
	searchResult := SearchResult{
		Name: tvMazeName,
	}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(tvMazeURL+"/search/shows?q=%s", q),
		nil)
	if err != nil {
		searchResult.Error = err
		return searchResult
	}

	result := &[]tvMazeResult{}
	err = GetJSON(req, result)
	if err != nil {
		searchResult.Error = err
		return searchResult
	}

	for _, r := range *result {
		searchResult.Shows = append(
			searchResult.Shows,
			Show{Title: r.Show.Title, ID: r.Show.ID, Source: tvMazeName})
	}

	return searchResult
}

func (t TvMaze) Seasons(show *store.Show) ([]Season, error) {
	return nil, nil
}

type tvMazeResult struct {
	Score float64    `json:"score"`
	Show  tvMazeShow `json:"show"`
}

type tvMazeShow struct {
	ID    int    `json:"id"`
	Title string `json:"name"`
}
