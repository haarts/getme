// Package sources provides the ability to search for shows/movies given a user
// provide search string.
package sources

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
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
	Title         string        `json:"title"`
	URL           string        `json:"url"`
	ID            int           `json:"id"`
	Seasons       []*Season     `json:"seasons"`
	SourceName    string        `json:"source_name"`
	IsDaily       bool          `json:"is_daily"`
	QuerySnippets QuerySnippets `json:"query_snippets"`
}

type QuerySnippets struct {
	ForEpisode []Snippet `json:"for_episode"`
	ForSeason  []Snippet `json:"for_season"`
}

type Snippet struct {
	Score         int    `json:"score"`
	TitleSnippet  string `json:"title_snippet"`
	FormatSnippet string `json:"format_snippet"`
}

// Season is _always_ part of a Show and contains meta data on a season in the show.
type Season struct {
	Season   int        `json:"season"`
	Episodes []*Episode `json:"episodes"`
}

// Episode is _always_ part of a Season and contains meta data on an episode in
// the show.
// TODO use TriedAt and Backoff to slowly stop trying to download episodes
// which prop never complete.
type Episode struct {
	Title   string    `json:"title"`
	Episode int       `json:"episode"`
	Pending bool      `json:"pending"`
	AirDate time.Time `json:"air_date"`
	season  int
	//TriedAt time.Time
	//Backoff int
}

func (e *Episode) Season() int {
	return e.season
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
		ms, err := s.Search(q) // TODO log.Error
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
		return fmt.Errorf("Source defined by show(%s) is not registered.", s.SourceName) // TODO log.Error
	}

	source := sources[s.SourceName]
	seasons, err := source.AllSeasonsAndEpisodes(*s) // pass a copy
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
func (s *Season) Done() {
	for _, episode := range s.Episodes {
		episode.Pending = false
	}
}

func (s *Show) BestSeasonSnippet() *Snippet {
	var best Snippet
	for _, snippet := range s.QuerySnippets.ForSeason {
		if snippet.Score >= best.Score {
			best = snippet
		}
	}
	return &best
}

func (s *Show) StoreSeasonSnippet(snippet Snippet) {
	for i, snip := range s.QuerySnippets.ForSeason {
		if snip.TitleSnippet == snippet.TitleSnippet && snip.FormatSnippet == snippet.FormatSnippet {
			s.QuerySnippets.ForSeason[i] = snippet
			return
		}
	}

	s.QuerySnippets.ForSeason = append(s.QuerySnippets.ForSeason, snippet)
}

func (s *Show) BestEpisodeSnippet() *Snippet {
	var best Snippet
	for _, snippet := range s.QuerySnippets.ForEpisode {
		if snippet.Score >= best.Score {
			best = snippet
		}
	}
	return &best
}

func (s *Show) StoreEpisodeSnippet(snippet Snippet) {
	for i, snip := range s.QuerySnippets.ForEpisode {
		if snip.TitleSnippet == snippet.TitleSnippet && snip.FormatSnippet == snippet.FormatSnippet {
			s.QuerySnippets.ForEpisode[i] = snippet
			return
		}
	}

	s.QuerySnippets.ForEpisode = append(s.QuerySnippets.ForEpisode, snippet)
}

// Done flags an episode as 'downloaded' and thus done. This episode is
// never looked up on a search engine agian.
func (e *Episode) Done() {
	e.Pending = false
}

// PendingSeasons return a list which is to be downloaded.
// A season is included when it is NOT the last season of the Show (high
// likelyhood of being still running and thus incomplete) and when all
// containing episodes are pending and it is NOT a daily show (people don't
// tend to bundle these).
func (s *Show) PendingSeasons() []*Season {
	var seasons []*Season
	for _, season := range s.Seasons {
		if s.isPending(season) {
			seasons = append(seasons, season)
		}
	}
	return seasons
}

// PendingEpisodes return a list which is to be downloaded.
func (s *Show) PendingEpisodes() []*Episode {
	var episodes []*Episode
	for _, season := range s.Seasons {
		if !s.isPending(season) {
			episodes = append(episodes, season.PendingEpisodes()...)
		}
	}
	return episodes
}

func (s *Show) isPending(season *Season) bool {
	if !s.IsDaily && !s.isLastSeason(season) && season.allEpisodesPending() {
		return true
	}
	return false
}

// PendingEpisodes returns all the episodes of this show which are still
// pending for download.
func (s *Season) PendingEpisodes() (episodes []*Episode) {
	for _, e := range s.Episodes {
		if e.Pending {
			e.season = s.Season
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

func getJSON(req *http.Request, target interface{}) error {
	return get(req, target, json.Unmarshal)
}

func getXML(req *http.Request, target interface{}) error {
	return get(req, target, xml.Unmarshal)
}

// get can be used to generically call URLs and deserialize the results.
func get(req *http.Request, target interface{}, unmarshalFunc func([]byte, interface{}) error) error {
	// TODO log.Debug(req)
	resp, err := http.DefaultClient.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return err //TODO retry a couple of times when it's a timeout.
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Search returned non 200 status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = unmarshalFunc(body, target)
	if err != nil {
		return err
	}

	return nil
}
