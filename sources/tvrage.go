package sources

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/haarts/getme/store"
)

// TvRage is the struct which implements the Source interface.
type TvRage struct{}

// TVRAGE defines the name of this source.
const tvRageName = "tvrage"

func init() {
	Register(tvRageName, TvRage{})
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

func tvRageRequest(URL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Search returns matches found by this source based on the query.
func (t TvRage) Search(query string) ([]Match, error) {
	req, err := tvRageRequest(constructTvRageSearchURL(query))

	if err != nil {
		return nil, err
	}

	result := &tvRageResult{}

	err = GetXML(req, result)
	if err != nil {
		return nil, err
	}

	shows := convertTvRageToMatches(result.Shows)
	putPopularShowOnTop(shows)

	return shows, nil
}

// AllSeasonsAndEpisodes finds the seasons and episodes for a show with this source.
func (t TvRage) AllSeasonsAndEpisodes(show store.Show) ([]*store.Season, error) {
	req, err := tvRageRequest(constructTvRageSeasonsURL(show.ID))
	if err != nil {
		return nil, err
	}

	result := &tvRageSeasonResult{}
	err = GetXML(req, result)
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
		matches[i] = store.Show{
			Title:      m.Title,
			ID:         m.ID,
			SourceName: tvRageName,
		}
	}
	return matches
}

// TODO Quite a bit of duplication with the convertToMatches function.
func convertFromTvRageSeasons(ss []tvRageSeason) []*store.Season {
	seasons := make([]*store.Season, len(ss))
	for i, s := range ss {
		season := &store.Season{
			Season:   s.Season,
			Episodes: make([]*store.Episode, len(s.Episodes)),
		}
		for j, e := range s.Episodes {
			season.Episodes[j] = &store.Episode{
				Title:   e.Title,
				Episode: e.Episode,
				Pending: true,
				AirDate: e.AirDate.Time,
			}
		}
		seasons[i] = season
	}
	return seasons
}
