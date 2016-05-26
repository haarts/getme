package sources

import (
	"fmt"
	"net/http"
	"time"

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

	// we assume the list is sorted...
	for _, r := range *result {
		ended := t.isEnded(r.Show.Status)
		searchResult.Shows = append(
			searchResult.Shows,
			Show{
				Title:  r.Show.Title,
				ID:     r.Show.ID,
				Ended:  &ended,
				URL:    r.Show.URL,
				Source: tvMazeName,
			})
	}

	return searchResult
}

// part of the value t because of namespace conflict.
func (t TvMaze) isEnded(status string) bool {
	return status == "Ended"
}

func (t TvMaze) Seasons(show *store.Show) ([]Season, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(tvMazeURL+"/shows/%d/episodes", show.ID),
		nil)
	if err != nil {
		return nil, err
	}
	result := &[]tvMazeEpisode{}
	err = GetJSON(req, result)
	if err != nil {
		return nil, err
	}

	seasons := map[int]*Season{}
	for _, r := range *result {
		if _, ok := seasons[r.Season]; !ok {
			seasons[r.Season] = &Season{Season: r.Season}
		}
		season := seasons[r.Season]
		if r.Airdate == nil {
			r.Airdate = &time.Time{}
		}
		season.Episodes = append(
			season.Episodes,
			Episode{Title: r.Name, Episode: r.Number, AirDate: *r.Airdate},
		)
	}

	s := []Season{}
	for _, v := range seasons {
		s = append(s, *v)
	}

	return s, nil
}

type tvMazeResult struct {
	Score float64    `json:"score"`
	Show  tvMazeShow `json:"show"`
}

type tvMazeShow struct {
	Title  string `json:"name"`
	ID     int    `json:"id"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

type tvMazeEpisode struct {
	Name    string     `json:"name"`
	Season  int        `json:"season"`
	Number  int        `json:"number"`
	Airdate *time.Time `json:"airstamp"`
}
