package store

import (
	"fmt"
	"math"
	"time"
)

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

// QuerySnippets is a collection of Snippets for episodes and seasons.
type QuerySnippets struct {
	ForEpisode []Snippet `json:"for_episode"`
	ForSeason  []Snippet `json:"for_season"`
}

// Snippet contains information on how a show can be found best. The
// TitleSnippet contains a possibly other name. The FormatSnippet contains how
// seasons/episodes are formatted.
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

type ByAirDate []*Episode

func (a ByAirDate) Len() int           { return len(a) }
func (a ByAirDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAirDate) Less(i, j int) bool { return a[i].AirDate.Unix() > a[j].AirDate.Unix() }

func (e *Episode) Season() int {
	return e.season
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
func (s *Show) DetermineIsDaily() bool {
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
