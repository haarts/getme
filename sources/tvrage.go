package sources

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func init() {
	Register("tvrage", searchTvRage)
}

type tvRageResult struct {
	Shows []tvRageMatch `xml:"show"`
}

type tvRageMatch struct {
	ID    int    `xml:"id"`
	Title string `xml:"name"`
}

type tvRageSeasonResult struct {
	EpisodeList tvRageEpisodeList `xml:"Episodelist"`
}

type tvRageEpisodeList struct {
	Seasons []tvRageSeason `xml:"Season"`
}

type tvRageSeason struct {
	Season   int             `xml:"attr,no"`
	Episodes []tvRageEpisode `xml:"episode"`
}

type tvRageEpisode struct {
	Episode int    `xml:"epnum"`
	Title   string `xml:"title"`
}

var tvRageURL = "http://services.tvrage.com"

func constructTvRageSearchURL(query string) string {
	return fmt.Sprintf(tvRageURL+"/feeds/search.php?show=%s", url.QueryEscape(query))
}

func constructTvRageSeasonsURL(show *Show) string {
	return fmt.Sprintf(tvRageURL+"/feeds/episode_list.php?sid=%d", show.ID)
}

func searchTvRage(query string) ([]Match, error) {
	resp, err := http.Get(constructTvRageSearchURL(query))
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

	var result tvRageResult
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	shows := convertTvRageToMatches(result.Shows)
	i := popularShowAtIndex(shows)
	if i != -1 {
		popularShow := shows[i]
		first := shows[0]
		shows[0] = popularShow
		shows[i] = first
	}

	return shows, nil
}

func convertTvRageToMatches(ms []tvRageMatch) []Match {
	matches := make([]Match, len(ms))
	for i, m := range ms {
		matches[i] = Show{
			Title: m.Title,
			ID:    m.ID,
			seasonsAndEpisodesFunc: getSeasonsOnTvRage,
		}
	}
	return matches
}

func getSeasonsOnTvRage(show *Show) error {
	resp, err := http.Get(constructTvRageSeasonsURL(show))
	if err != nil {
		return err //TODO retry a couple of times when it's a timeout.
	}
	if resp.StatusCode != 200 {
		return errors.New("Search return non 200 status code")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result tvRageSeasonResult
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	show.Seasons = convertFromTvRageSeasons(show, result.EpisodeList.Seasons)
	return nil
}

func convertFromTvRageSeasons(show *Show, ss []tvRageSeason) []*Season {
	seasons := make([]*Season, len(ss))
	for i, s := range ss {
		season := Season{
			Episodes: make([]*Episode, len(s.Episodes)),
		}
		for i, e := range s.Episodes {
			season.Episodes[i] = &Episode{e.Title, &season, e.Episode}
		}
		seasons[i] = &season
	}
	return seasons
}
