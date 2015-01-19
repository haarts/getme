// Package sources provides the ability to search for shows/movies given a user
// provide search string.
package sources

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/haarts/getme/config"
	"github.com/haarts/getme/store"
)

var conf = config.Config()
var log = config.Log()

// Source defines the methods a new source for media info should implement.
type Source interface {
	Search(string) ([]Match, error)
	AllSeasonsAndEpisodes(store.Show) ([]*store.Season, error)
}

// Match is either a Movie or a Show.
type Match interface {
	DisplayTitle() string
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

// TODO Done I need this?
func findSource(name string) (Source, error) {
	source, ok := sources[name]
	if !ok {
		//log.WithFields(
		//logrus.Fields{
		//"source": source.SourceName,
		//"show":   source.Title,
		//}).Error("Source defined by show not registered")
		return nil, fmt.Errorf("Source defined by show is not registered.") //, s.SourceName)
	}

	return source, nil
}

// GetSeasonsAndEpisodes should be called with a newly initialized Show. In
// addition to fetching the seasons and episodes this determines if the show is
// a daily show, which impacts the way queries are generated for search
// engines.
func GetSeasonsAndEpisodes(s *store.Show) error {
	err := UpdateSeasonsAndEpisodes(s)
	if err != nil {
		return err
	}

	s.IsDaily = s.DetermineIsDaily()
	return nil
}

// UpdateSeasonsAndEpisodes should be called to update a Show after, for
// example, deserialization for disk.
func UpdateSeasonsAndEpisodes(s *store.Show) error {
	source, err := findSource(s.SourceName)
	if err != nil {
		return err
	}
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

func addSeason(show *store.Show, season *store.Season) {
	show.Seasons = append(show.Seasons, season)
}

func updateEpisodes(existingSeason *store.Season, newSeason *store.Season) {
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

func contains(ss []*store.Episode, e *store.Episode) bool {
	for _, a := range ss {
		if a.Episode == e.Episode {
			return true
		}
	}
	return false
}

func findExistingSeason(existing []*store.Season, other *store.Season) *store.Season {
	for _, season := range existing {
		if season.Season == other.Season {
			return season
		}
	}
	return nil
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
	for name, s := range sources { //TODO Make parallel
		ms, err := s.Search(q)
		if err != nil {
			log.WithFields(
				logrus.Fields{
					"error":  err,
					"source": name,
				}).Error("Error when searching on source")
		}
		matches[i] = ms
		errors = append(errors, err)
		i++
	}
	return matches, errors
}
