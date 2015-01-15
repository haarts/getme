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

func (t Trakt) seasonURL(show Show, season Season) string {
	return fmt.Sprintf(traktURL+"/shows/%s/seasons/%d?extended=full", show.URL, season.Season)
}

func (t Trakt) seasonsURL(show Show) string {
	return fmt.Sprintf(traktURL+"/shows/%s/seasons?extended=full", show.URL)
}

// AllSeasonsAndEpisodes finds the seasons and episodes for a show with this source.
func (t Trakt) AllSeasonsAndEpisodes(show Show) ([]*Season, error) {
	req, err := traktRequest(t.seasonsURL(show))
	if err != nil {
		return nil, err
	}

	ss := &([]traktSeason{})

	err = getJSON(req, ss)
	if err != nil {
		return nil, err
	}

	seasons := convertToSeasons(*ss)
	err = t.addEpisodes(seasons, show)
	if err != nil {
		return nil, err
	}
	return seasons, nil
}

// TODO Quite a bit of duplication with the convertToMatches function.
func convertToSeasons(ss []traktSeason) []*Season {
	seasons := make([]*Season, 0, len(ss))
	for _, s := range ss {
		season := &Season{
			Season:   s.Season,
			Episodes: make([]*Episode, 0, s.Episodes),
		}
		seasons = append(seasons, season)
	}
	return seasons
}

func (t Trakt) addEpisodes(seasons []*Season, show Show) error {
	for _, season := range seasons {
		req, err := traktRequest(t.seasonURL(show, *season))
		if err != nil {
			return err
		}

		episodes := &([]traktEpisode{})
		err = getJSON(req, episodes)
		if err != nil {
			return err
		}

		for _, episode := range *episodes {
			if episode.FirstAired == nil {
				episode.FirstAired = &time.Time{}
			}
			e := Episode{
				Title:   episode.Title,
				Episode: episode.Number,
				Pending: true, // NOTE Do not forget to set pending to true!
				AirDate: *episode.FirstAired,
			}

			season.Episodes = append(season.Episodes, &e)
		}
	}
	return nil
}
