package sources

import "fmt"

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
	Title   string
	URL     string
	Seasons []*Season
}

func (s Show) DisplayTitle() string {
	return s.Title
}

func (s Show) Episodes() []Episode {
	return nil
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
}

type searchFun func(string) ([]Match, error)

var sources = make(map[string]searchFun)

func Register(name string, source searchFun) {
	if _, dup := sources[name]; dup {
		panic("source: Register called twice for source " + name)
	}
	sources[name] = source
}

func (e *Episode) ShowName() string {
	return e.Season.Show.Title
}

func (e *Episode) QueryNames() []string {
	// TODO deal with daily shows
	s1 := fmt.Sprintf("%s S%02dE%02d", e.ShowName(), e.Season.Season, e.Episode)
	s2 := fmt.Sprintf("%s %dx%d", e.ShowName(), e.Season.Season, e.Episode)

	return []string{s1, s2}
}
