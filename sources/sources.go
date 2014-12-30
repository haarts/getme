package sources

import (
	"fmt"
	"regexp"
	"time"
)

type Match interface {
	DisplayTitle() string
}

type Movie struct {
	Title string
}

type Show struct {
	Title                  string    `json:"title"`
	URL                    string    `json:"url"`
	ID                     int       `json:"id"`
	Seasons                []*Season `json:""`
	seasonsAndEpisodesFunc func(*Show) error
	isDaily                bool `json:"is_daily"`
}

type Season struct {
	Show     *Show
	Season   int        `json:"season"`
	Episodes []*Episode `json:"episodes"`
}

// TODO use TriedAt and Backoff to slowly stop trying to download episodes which prop never complete.
type Episode struct {
	Season  *Season
	Title   string    `json:"title"`
	Episode int       `json:"episode"`
	Pending bool      `json:"pending"`
	AirDate time.Time `json:"air_date"`
	//TriedAt time.Time
	//Backoff int
}

type searchFun func(string) ([]Match, error)

var sources = make(map[string]searchFun)

func Register(name string, source searchFun) {
	if _, dup := sources[name]; dup {
		panic("source: Register called twice for source " + name)
	}
	sources[name] = source
}

func ListSources() (names []string) {
	for name, _ := range sources {
		names = append(names, name)
	}
	return
}

func Search(q string) ([][]Match, []error) {
	var matches = make([][]Match, len(sources))
	var errors = make([]error, len(sources))
	var i int
	for _, s := range sources { //TODO Make parallel
		ms, err := s(q)
		matches[i] = ms
		errors = append(errors, err)
		i++
	}
	return matches, errors
}

func (m Movie) DisplayTitle() string {
	return m.Title
}

func (s *Show) String() string {
	return fmt.Sprintf(
		"Show: %s, number of Seasons: %d, number of episodes: %d",
		s.Title,
		len(s.Seasons),
		len(s.Episodes()),
	)
}

func (s Show) DisplayTitle() string {
	return s.Title
}

func (s *Show) GetSeasonsAndEpisodes() error {
	err := s.seasonsAndEpisodesFunc(s)
	s.isDaily = s.determineIsDaily()
	return err
}

func (s Show) Episodes() (episodes []*Episode) {
	for _, season := range s.Seasons {
		episodes = append(episodes, season.Episodes...)
	}
	return
}

func (s Show) PendingEpisodes() (episodes []*Episode) {
	allEpisodes := s.Episodes()
	for _, e := range allEpisodes {
		if e.Pending {
			episodes = append(episodes, e)
		}
	}
	return
}

func (s *Season) AllEpisodesPending() bool {
	for _, e := range s.Episodes {
		if !e.Pending {
			return false
		}
	}
	return true
}

func (e *Episode) ShowName() string {
	return e.Season.Show.Title
}

func (e *Episode) AsFileName() string {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	return string(re.ReplaceAll([]byte(e.QueryNames()[0]), []byte("_")))
}

// This is a heuristic really.
func (s *Show) determineIsDaily() bool {
	season := s.Seasons[0]
	// If there are more than 30 episodes in a season it MIGHT be a daily.
	if len(season.Episodes) > 30 {

		// And if there are two episodes with consecutive AirDate's it's PROB a daily.
		d1 := season.Episodes[4].AirDate // Early in the season, too early for breaks.
		d2 := season.Episodes[5].AirDate

		// Without an airdate we can't say.
		if d1.IsZero() || d2.IsZero() {
			return false
		}
		if isNextDay(d1, d2) {
			return true
		}
	}
	return false
}

func isNextDay(d1, d2 time.Time) bool {
	d := d1.Sub(d2)
	return d.Hours()/24 == -1
}

func (e *Episode) QueryNames() []string {
	if e.Season.Show.isDaily { // Potential train wreck
		y, m, d := e.AirDate.Date()
		return []string{fmt.Sprintf("%s %d %d %d", e.ShowName(), y, m, d)}
	} else {
		s1 := fmt.Sprintf("%s S%02dE%02d", e.ShowName(), e.Season.Season, e.Episode)
		s2 := fmt.Sprintf("%s %dx%d", e.ShowName(), e.Season.Season, e.Episode)
		return []string{s1, s2}
	}
}
