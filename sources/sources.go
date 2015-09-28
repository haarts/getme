// Package sources provides the ability to search for shows/movies given a user
// provide search string.
package sources

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/store"
)

type Source interface {
	Search(string) SearchResult
	Seasons(*store.Show) ([]Season, error)
	Name() string
}

type Match interface {
	DisplayTitle() string
}

type Show struct {
	Title  string
	ID     int
	Source string
}

func (s Show) DisplayTitle() string {
	return s.Title
}

type Season struct {
	Season   int
	Episodes []Episode
}

type Episode struct {
	Title   string    `json:"title"`
	Episode int       `json:"episode"`
	AirDate time.Time `json:"air_date"`
}

// SearchResult holds the results of searching on a particular source for a
// particular query.
type SearchResult struct {
	Name  string
	Shows []Show
	Error error
}

// sources contains all sources one can query for show information
var sources = map[string]Source{
	Trakt{}.Name():  Trakt{},
	TvRage{}.Name(): TvRage{},
	TvMaze{}.Name(): TvMaze{},
}

func SourceNames() (names []string) {
	for k := range sources {
		names = append(names, k)
	}
	return
}

// UpdateSeasonsAndEpisodes should be called to update a Show after, for
// example, deserialization from disk.
func UpdateSeasonsAndEpisodes(show *store.Show) error {
	var seasons []Season
	var err error

	seasons, err = sources[show.SourceName].Seasons(show)
	if err != nil {
		return err
	}

	for i := 0; i < len(seasons); i++ {
		season := seasons[i]
		existingSeason := findExistingSeason(show.Seasons, season)
		if existingSeason == nil {
			addSeason(show, season)
		} else {
			updateEpisodes(existingSeason, season)
		}
	}

	return nil
}

// Search is the important function of this package. Call this to turn a user
// search string into a list of matches (which might be TV shows or movies).
func Search(q string) []SearchResult {
	c := make(chan SearchResult)
	for _, source := range sources {
		go func(s Source) { c <- s.Search(q) }(source)
	}

	var searchResults []SearchResult
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(sources); i++ {
		select {
		case result := <-c:
			searchResults = append(searchResults, result)
		case <-timeout:
			log.Error("Search timed out")
			return searchResults
		}
	}

	return searchResults
}

func addSeason(show *store.Show, season Season) {
	newSeason := store.Season{
		Season: season.Season,
	}
	for _, episode := range season.Episodes {
		newEpisode := store.Episode{
			Episode: episode.Episode,
			AirDate: episode.AirDate,
			Title:   episode.Title,
			Pending: true,
		}
		newSeason.Episodes = append(newSeason.Episodes, &newEpisode)
	}

	show.Seasons = append(show.Seasons, &newSeason)
}

func updateEpisodes(existingSeason *store.Season, newSeason Season) {
	for i, e := range existingSeason.Episodes { // Delete 'TBA' episodes.
		if e.Title == "TBA" {
			existingSeason.Episodes[i] = nil
			existingSeason.Episodes =
				append(existingSeason.Episodes[:i], existingSeason.Episodes[i+1:]...)
		}
	}

	if len(existingSeason.Episodes) == len(newSeason.Episodes) {
		return
	}

	for _, episode := range newSeason.Episodes {
		if !contains(existingSeason.Episodes, episode) {
			newEpisode := store.Episode{
				Episode: episode.Episode,
				AirDate: episode.AirDate,
				Title:   episode.Title,
				Pending: true,
			}
			existingSeason.Episodes = append(existingSeason.Episodes, &newEpisode)
		}
	}
}

func contains(episodes []*store.Episode, other Episode) bool {
	for _, e := range episodes {
		if e.Episode == other.Episode {
			return true
		}
	}
	return false
}

func findExistingSeason(existing []*store.Season, other Season) *store.Season {
	for _, season := range existing {
		if season.Season == other.Season {
			return season
		}
	}
	return nil
}
