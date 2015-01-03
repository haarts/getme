package sources

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type traktSeason struct {
	Season   int `json:"season"`
	Episodes int `json:"episodes"`
}

var traktSeasonsURL = "http://api.trakt.tv/show/seasons.json/5bc6254d3bbde304a49557cf2845d921/"

func (t Trakt) AllSeasonsAndEpisodes(show Show) ([]*Season, error) {
	parts := strings.Split(show.URL, "/")
	traktIdentifier := parts[len(parts)-1]

	resp, err := http.Get(traktSeasonsURL + traktIdentifier)
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
		for j, _ := range season.Episodes {
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
