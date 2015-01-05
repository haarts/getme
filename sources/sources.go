package sources

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

type Source interface {
	Search(string) ([]Match, error)
	AllSeasonsAndEpisodes(Show) ([]*Season, error)
}

type Match interface {
	DisplayTitle() string
}

type Movie struct {
	Title string
}

type Show struct {
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	ID         int       `json:"id"`
	Seasons    []*Season `json:"seasons"`
	SourceName string    `json:"source_name"`
	isDaily    bool      `json:"is_daily"`
}

type Season struct {
	Show     *Show      `json:"-"`
	Season   int        `json:"season"`
	Episodes []*Episode `json:"episodes"`
}

// TODO use TriedAt and Backoff to slowly stop trying to download episodes which prop never complete.
type Episode struct {
	Season  *Season   `json:"-"`
	Title   string    `json:"title"`
	Episode int       `json:"episode"`
	Pending bool      `json:"pending"`
	AirDate time.Time `json:"air_date"`
	//TriedAt time.Time
	//Backoff int
}

type PendingItem struct {
	//NOTE Perhaps we could return a set of fields in the future and let the search
	//engine construct the string they want,
	QueryNames []string
	ShowTitle  string
	episodes   []*Episode
}

var sources = make(map[string]Source)

func Register(name string, source Source) {
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
		ms, err := s.Search(q)
		matches[i] = ms
		errors = append(errors, err)
		i++
	}
	return matches, errors
}

// Initial adding of episodes.
func GetSeasonsAndEpisodes(s *Show) error {
	err := UpdateSeasonsAndEpisodes(s)
	if err != nil {
		return err
	}

	s.isDaily = s.determineIsDaily()
	return nil
}

// Subsequent updates of episodes
func UpdateSeasonsAndEpisodes(s *Show) error {
	if _, ok := sources[s.SourceName]; !ok {
		return errors.New(fmt.Sprintf("Source defined by show(%s) is not registered.", s.SourceName))
	}

	source := sources[s.SourceName]
	seasons, err := source.AllSeasonsAndEpisodes(*s)
	if err != nil {
		return err
	}

	for i := 0; i < len(seasons); i++ {
		season := seasons[i]
		existingSeason := findExistingSeason(s.Seasons, season)
		if existingSeason == nil {
			addSeason(s, season)
		} else {
			updateEpisodes(existingSeason, season)
		}
	}

	return nil
}

func addSeason(show *Show, season *Season) {
	season.Show = show
	show.Seasons = append(show.Seasons, season)
}

func updateEpisodes(existingSeason *Season, newSeason *Season) {
	if len(existingSeason.Episodes) == len(newSeason.Episodes) {
		return
	}

	for _, episode := range newSeason.Episodes {
		if !contains(existingSeason.Episodes, episode) {
			newEpisode := *episode
			newEpisode.Season = existingSeason
			existingSeason.Episodes = append(existingSeason.Episodes, &newEpisode)
		}
	}
}

func contains(ss []*Episode, e *Episode) bool {
	for _, a := range ss {
		if a.Episode == e.Episode {
			return true
		}
	}
	return false
}

func findExistingSeason(existing []*Season, other *Season) *Season {
	for _, season := range existing {
		if season.Season == other.Season {
			return season
		}
	}
	return nil
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

func (s Show) Episodes() (episodes []*Episode) {
	for _, season := range s.Seasons {
		episodes = append(episodes, season.Episodes...)
	}
	return
}

func (p *PendingItem) Done() {
	for _, episode := range p.episodes {
		episode.Pending = false
	}
}

// TODO cut this method up.
// PendingItems returns a list of items which are to be downloaded.
// The list consists of a mix of episodes and seasons.
// A season is included when it is NOT the last season of the Show and when all containing episodes
// are pending.
func (show Show) PendingItems() (pendingItems []PendingItem) {
	for _, season := range show.Seasons {
		if !show.isLastSeason(season) && season.allEpisodesPending() {
			item := PendingItem{
				QueryNames: []string{fmt.Sprintf("%s season %d", show.Title, season.Season)},
				episodes:   season.Episodes,
				ShowTitle:  show.Title,
			}
			pendingItems = append(pendingItems, item)
		} else {
			episodes := season.PendingEpisodes()
			for _, episode := range episodes {
				item := PendingItem{
					episodes:  []*Episode{episode},
					ShowTitle: show.Title,
				}
				if show.isDaily {
					y, m, d := episode.AirDate.Date()
					item.QueryNames = []string{fmt.Sprintf("%s %d %d %d", show.Title, y, m, d)}
				} else {
					s1 := fmt.Sprintf("%s S%02dE%02d", show.Title, season.Season, episode.Episode)
					s2 := fmt.Sprintf("%s %dx%d", show.Title, season.Season, episode.Episode)
					item.QueryNames = []string{s1, s2}
				}
				pendingItems = append(pendingItems, item)
			}
		}
	}
	return
}

func (season Season) PendingEpisodes() (episodes []*Episode) {
	for _, e := range season.Episodes {
		if e.Pending {
			episodes = append(episodes, e)
		}
	}
	return
}

// TODO deprecated
func (s Show) PendingEpisodes() (episodes []*Episode) {
	allEpisodes := s.Episodes()
	for _, e := range allEpisodes {
		if e.Pending {
			episodes = append(episodes, e)
		}
	}
	return
}

func (s *Season) allEpisodesPending() bool {
	for _, e := range s.Episodes {
		if !e.Pending {
			return false
		}
	}
	return true
}

// NOTE Can't assume ordering.
func (show Show) isLastSeason(season *Season) bool {
	for _, s := range show.Seasons {
		if s.Season > season.Season {
			return false
		}
	}
	return true
}

// TODO deprecated?
func (e *Episode) ShowName() string {
	return e.Season.Show.Title
}

// TODO deprecated?
func (e *Episode) AsFileName() string {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	return string(re.ReplaceAll([]byte(e.QueryNames()[0]), []byte("_")))
}

// NOTE This is a heuristic really.
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

// TODO deprecated
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
