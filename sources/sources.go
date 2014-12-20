package sources

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
	Season   int
	Episodes int
}

type Episode struct {
	Title   string
	Season  int
	Episode int
}

func CreateEpisodes(seasons []Season) (episodes []Episode) {
	for _, s := range seasons {
		for i := 1; i <= s.Episodes; i++ {
			episodes = append(episodes, Episode{Season: s.Season, Episode: i})
		}
	}
	return
}
