package sources

import "fmt"

type Match struct {
	Title string
	URL   string // NOTE: Perhaps this should be a url.URL in stead of a string
}

type Movie struct {
}

type Show struct {
	Title string
}

type Season struct {
	ShowName string
	Season   int
	Episodes int
}

type Episode struct {
	ShowName string
	Title    string
	Season   int
	Episode  int
}

func CreateEpisodes(seasons []Season) (episodes []Episode) {
	for _, s := range seasons {
		for i := 1; i <= s.Episodes; i++ {
			episodes = append(episodes, Episode{ShowName: s.ShowName, Season: s.Season, Episode: i})
		}
	}
	return
}

func (e *Episode) QueryNames() []string {
	// TODO deal with daily shows
	s1 := fmt.Sprintf("%s S%02dE%02d", e.ShowName, e.Season, e.Episode)
	s2 := fmt.Sprintf("%s %dx%d", e.ShowName, e.Season, e.Episode)

	return []string{s1, s2}
}
