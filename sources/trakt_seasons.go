package sources

import (
	"fmt"
	"time"
)

type traktSeason struct {
	Season   int `json:"number"`
	Episodes int `json:"episode_count"`
}

type traktEpisode struct {
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	FirstAired *time.Time `json:"first_aired"`
}

var traktSeasonsURL = traktURL + "/shows/%s/seasons?extended=full"
var traktSeasonURL = traktURL + "/shows/%s/seasons/%d?extended=full"

// AllSeasonsAndEpisodes finds the seasons and episodes for a show with this source.
func (t Trakt) AllSeasonsAndEpisodes(show Show) ([]*Season, error) {
	req, err := traktRequest(fmt.Sprintf(traktSeasonsURL, show.URL))
	if err != nil {
		return nil, err
	}

	ss := &([]traktSeason{})

	err = get(req, ss)
	if err != nil {
		return nil, err
	}

	seasons := convertToSeasons(*ss)
	err = addEpisodes(seasons, show)
	if err != nil {
		return nil, err
	}
	return seasons, nil
}

// TODO Quite a bit of duplication with the convertToMatches function.
func convertToSeasons(ss []traktSeason) []*Season {
	seasons := make([]*Season, len(ss))
	for i, s := range ss {
		season := &Season{
			Season:   s.Season,
			Episodes: make([]*Episode, s.Episodes),
		}
		seasons[i] = season
	}
	return seasons
}

func addEpisodes(seasons []*Season, show Show) error {
	for _, season := range seasons {
		req, err := traktRequest(fmt.Sprintf(traktSeasonURL, show.URL, season.Season))
		if err != nil {
			return err
		}

		episodes := &([]traktEpisode{})
		err = get(req, episodes)
		if err != nil {
			return err
		}

		for i, episode := range *episodes {
			if episode.FirstAired == nil {
				episode.FirstAired = &time.Time{}
			}
			season.Episodes[i] = &Episode{
				Title:   episode.Title,
				Episode: episode.Number,
				Pending: true, // NOTE Do not forget to set pending to true!
				AirDate: *episode.FirstAired,
			}
		}
	}
	return nil
}
