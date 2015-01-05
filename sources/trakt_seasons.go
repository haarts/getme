package sources

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type traktSeason struct {
	Season   int `json:"number"`
	Episodes int `json:"episode_count"`
}

var traktSeasonsURL = traktURL + "/shows/%s/seasons?extended=full"

// AllSeasonsAndEpisodes finds the seasons and episodes for a show with this source.
func (t Trakt) AllSeasonsAndEpisodes(show Show) ([]*Season, error) {
	req, err := traktRequest(fmt.Sprintf(traktSeasonsURL, show.URL))
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
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

	var ss []traktSeason
	err = json.Unmarshal(body, &ss)
	if err != nil {
		return nil, err
	}

	return convertToSeasons(ss), nil
}

// TODO Quite a bit of duplication with the convertToMatches function.
func convertToSeasons(ss []traktSeason) []*Season {
	seasons := make([]*Season, len(ss))
	for i, s := range ss {
		season := &Season{
			Season:   s.Season,
			Episodes: make([]*Episode, s.Episodes),
		}
		for j := range season.Episodes {
			season.Episodes[j] = &Episode{
				Title:   "",
				Season:  season,
				Episode: j + 1,
				Pending: true, // NOTE Do not forget to set pending to true!
				AirDate: time.Time{},
			}
		}
		seasons[i] = season
	}
	return seasons
}
