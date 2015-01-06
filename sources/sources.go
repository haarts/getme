// Package sources provides the ability to search for shows/movies given a user
// provide search string.
package sources

import (
	"fmt"
	"math"
	"time"
)

// Source defines the methods a new source for media info should implement.
type Source interface {
	Search(string) ([]Match, error)
	AllSeasonsAndEpisodes(Show) ([]*Season, error)
}

// Match is either a Movie or a Show.
type Match interface {
	DisplayTitle() string
}

// Movie contains all the relevant information for a movie.
type Movie struct {
	Title string
}

// Show contains all the relevant information for a TV show. A value is Show is
// the main way on interfacing with the show, seasons AND episodes.
type Show struct {
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	ID         int       `json:"id"`
	Seasons    []*Season `json:"seasons"`
	SourceName string    `json:"source_name"`
	IsDaily    bool      `json:"is_daily"`
}

// Season is _always_ part of a Show and contains meta data on a season in the show.
type Season struct {
	Season   int        `json:"season"`
	Episodes []*Episode `json:"episodes"`
}

// Episode is _always_ part of a Season and contains meta data on an episode in the show.
// TODO use TriedAt and Backoff to slowly stop trying to download episodes which prop never complete.
type Episode struct {
	Title   string    `json:"title"`
	Episode int       `json:"episode"`
	Pending bool      `json:"pending"`
	AirDate time.Time `json:"air_date"`
	//TriedAt time.Time
	//Backoff int
}

// PendingItem is a generalisation of something which hasn't been downloaded
// yet. This can be a single episode OR an entire season.
type PendingItem struct {
	//NOTE Perhaps we could return a set of fields in the future and let the search
	//engine construct the string they want,
	QueryNames []string
	ShowTitle  string
	episodes   []*Episode
}

var sources = make(map[string]Source)

// Register should be called (in an init function) when adding a new source.
// See trakt.go for an example.
func Register(name string, source Source) {
	if _, dup := sources[name]; dup {
		panic("source: Register called twice for source " + name)
	}
	sources[name] = source
}

// ListSources exists mainly for display purposes.
func ListSources() (names []string) {
	for name := range sources {
		names = append(names, name)
	}
	return
}

// Search is the important function of this package. Call this to turn a user
// search string into a list of matches (which might be TV shows or movies).
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

// GetSeasonsAndEpisodes should be called with a newly initialized Show. In
// addition to fetching the seasons and episodes this determines if the show is
// a daily show, which impacts the way queries are generated for search
// engines.
func GetSeasonsAndEpisodes(s *Show) error {
	err := UpdateSeasonsAndEpisodes(s)
	if err != nil {
		return err
	}

	s.IsDaily = s.determineIsDaily()
	return nil
}

// UpdateSeasonsAndEpisodes should be called to update a Show after, for
// example, deserialization for disk.
func UpdateSeasonsAndEpisodes(s *Show) error {
	if _, ok := sources[s.SourceName]; !ok {
		return fmt.Errorf("Source defined by show(%s) is not registered.", s.SourceName)
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
	show.Seasons = append(show.Seasons, season)
}

func updateEpisodes(existingSeason *Season, newSeason *Season) {
	if len(existingSeason.Episodes) == len(newSeason.Episodes) {
		return
	}

	for _, episode := range newSeason.Episodes {
		if !contains(existingSeason.Episodes, episode) {
			newEpisode := *episode
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

// DisplayTitle returns the title of a movie. Here to satisfy the Match
// interface.
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

// DisplayTitle returns the title of a TV show. Here to satisfy the Match
// interface.
func (s Show) DisplayTitle() string {
	return s.Title
}

// Episodes lists all the episodes associated with this show.
func (s Show) Episodes() (episodes []*Episode) {
	for _, season := range s.Seasons {
		episodes = append(episodes, season.Episodes...)
	}
	return
}

// Done flags an episode as 'downloaded' and thus done. This episode is
// never looked up on a search engine agian.
func (p *PendingItem) Done() {
	for _, episode := range p.episodes {
		episode.Pending = false
	}
}

// PendingItems returns a list of items which are to be downloaded.
// The list consists of a mix of episodes and seasons.
// A season is included when it is NOT the last season of the Show and when all containing episodes
// are pending and it is NOT a daily show.
// TODO cut this method up.
func (s Show) PendingItems() (pendingItems []PendingItem) {
	for _, season := range s.Seasons {
		if !s.IsDaily && !s.isLastSeason(season) && season.allEpisodesPending() {
			pendingItems = append(pendingItems, itemForEntireSeason(s, season))
		} else {
			episodes := season.PendingEpisodes()
			for _, episode := range episodes {
				item := itemForEpisode(s, season, episode)
				pendingItems = append(pendingItems, item)
			}
		}
	}
	return
}

func itemForEpisode(show Show, season *Season, episode *Episode) PendingItem {
	item := PendingItem{
		episodes:  []*Episode{episode},
		ShowTitle: show.Title,
	}

	if show.IsDaily {
		y, m, d := episode.AirDate.Date()
		item.QueryNames = []string{fmt.Sprintf("%s %d %d %02d", show.Title, y, m, d)}
	} else {
		s1 := fmt.Sprintf("%s S%02dE%02d", show.Title, season.Season, episode.Episode)
		s2 := fmt.Sprintf("%s %dx%d", show.Title, season.Season, episode.Episode)
		item.QueryNames = []string{s1, s2}
	}

	return item
}

func itemForEntireSeason(show Show, season *Season) PendingItem {
	return PendingItem{
		QueryNames: []string{fmt.Sprintf("%s season %d", show.Title, season.Season)},
		episodes:   season.Episodes,
		ShowTitle:  show.Title,
	}
}

// PendingEpisodes returns all the episodes of this show which are still
// pending for download.
func (season Season) PendingEpisodes() (episodes []*Episode) {
	for _, e := range season.Episodes {
		if e.Pending {
			episodes = append(episodes, e)
		}
	}
	return
}

func (season *Season) allEpisodesPending() bool {
	for _, e := range season.Episodes {
		if !e.Pending {
			return false
		}
	}
	return true
}

// NOTE Can't assume ordering.
func (s Show) isLastSeason(currentSeason *Season) bool {
	for _, season := range s.Seasons {
		if season.Season > currentSeason.Season {
			return false
		}
	}
	return true
}

// NOTE This is a heuristic really.
func (s *Show) determineIsDaily() bool {
	// Prefer the second to last season. If not there get the first.
	season := s.Seasons[int(math.Max(0, float64(len(s.Seasons)-2)))]
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
