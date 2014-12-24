package sources

import (
	"fmt"
	"regexp"
)

type Match interface {
	DisplayTitle() string
}

type Movie struct {
	Title string
}

func (m Movie) DisplayTitle() string {
	return m.Title
}

type Show struct {
	Title                  string
	URL                    string
	ID                     int
	Seasons                []*Season
	seasonsAndEpisodesFunc func(*Show) error
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
	return s.seasonsAndEpisodesFunc(s)
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

type Season struct {
	Show     *Show
	Season   int
	Episodes []*Episode
}

type Episode struct {
	Title   string
	Season  *Season
	Episode int
	Pending bool
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

func Search(q string) (matches []Match, errors []error) {
	for _, s := range sources { //TODO Make parallel
		ms, err := s(q)
		matches = append(matches, ms...) // TODO This is pretty dumb of course, with multiple sources the first always wins.
		errors = append(errors, err)
	}
	return
}

func (e *Episode) ShowName() string {
	return e.Season.Show.Title
}

func (e *Episode) AsFileName() string {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	return string(re.ReplaceAll([]byte(e.QueryNames()[0]), []byte("_")))
}

func (e *Episode) QueryNames() []string {
	// TODO deal with daily shows
	s1 := fmt.Sprintf("%s S%02dE%02d", e.ShowName(), e.Season.Season, e.Episode)
	s2 := fmt.Sprintf("%s %dx%d", e.ShowName(), e.Season.Season, e.Episode)

	return []string{s1, s2}
}
