package sources

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// TvRage is the struct which implements the Source interface.
type TvRage struct{}

// TVRAGE defines the name of this source.
const TVRAGE = "tvrage"

func init() {
	Register(TVRAGE, TvRage{})
}

type tvRageResult struct {
	Shows []tvRageMatch `xml:"show"`
}

type tvRageMatch struct {
	ID    int    `xml:"showid"`
	Title string `xml:"name"`
}

type tvRageSeasonResult struct {
	EpisodeList tvRageEpisodeList `xml:"Episodelist"`
}

type tvRageEpisodeList struct {
	Seasons []tvRageSeason `xml:"Season"`
}

type tvRageSeason struct {
	Season   int             `xml:"no,attr"`
	Episodes []tvRageEpisode `xml:"episode"`
}

type tvRageEpisode struct {
	Episode int        `xml:"seasonnum"`
	Title   string     `xml:"title"`
	AirDate tvRageDate `xml:"airdate"`
}

type tvRageDate struct {
	time.Time
}

var tvRageURL = "http://services.tvrage.com"

// Search returns matches found by this source based on the query.
func (t TvRage) Search(query string) ([]Match, error) {
	resp, err := http.Get(constructTvRageSearchURL(query))
	if err != nil {
		return nil, err //TODO retry a couple of times when it's a timeout.
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Search returned non 200 status code: %d", resp.StatusCode)
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
	putPopularShowOnTop(shows)

	return shows, nil
}

// AllSeasonsAndEpisodes finds the seasons and episodes for a show with this source.
func (t TvRage) AllSeasonsAndEpisodes(show Show) ([]*Season, error) {
	resp, err := http.Get(constructTvRageSeasonsURL(show.ID))
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

	var result tvRageSeasonResult
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return convertFromTvRageSeasons(result.EpisodeList.Seasons), nil
}

func (t *tvRageDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	const dateFormat = "2006-01-02"
	var s string
	d.DecodeElement(&s, &start)
	parse, err := time.Parse(dateFormat, s)
	if err != nil {
		return nil
	}
	*t = tvRageDate{parse}
	return nil
}

func constructTvRageSearchURL(query string) string {
	return fmt.Sprintf(tvRageURL+"/feeds/search.php?show=%s", url.QueryEscape(query))
}

func constructTvRageSeasonsURL(ID int) string {
	return fmt.Sprintf(tvRageURL+"/feeds/episode_list.php?sid=%d", ID)
}

func putPopularShowOnTop(shows []Match) {
	i := popularShowAtIndex(shows)
	if i != -1 {
		popularShow := shows[i]
		first := shows[0]
		shows[0] = popularShow
		shows[i] = first
	}
}

func convertTvRageToMatches(ms []tvRageMatch) []Match {
	matches := make([]Match, len(ms))
	for i, m := range ms {
		matches[i] = Show{
			Title:      m.Title,
			ID:         m.ID,
			SourceName: TVRAGE,
		}
	}
	return matches
}

// TODO Quite a bit of duplication with the convertToMatches function.
func convertFromTvRageSeasons(ss []tvRageSeason) []*Season {
	seasons := make([]*Season, len(ss))
	for i, s := range ss {
		season := &Season{
			Season:   s.Season,
			Episodes: make([]*Episode, len(s.Episodes)),
		}
		for j, e := range s.Episodes {
			season.Episodes[j] = &Episode{
				Title:   e.Title,
				Season:  season,
				Episode: e.Episode,
				Pending: true,
				AirDate: e.AirDate.Time,
			}
		}
		seasons[i] = season
	}
	return seasons
}
