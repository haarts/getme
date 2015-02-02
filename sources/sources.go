// Package sources provides the ability to search for shows/movies given a user
// provide search string.
package sources

import (
	"time"

	trakt "github.com/42minutes/go-trakt"

	"github.com/haarts/getme/store"
)

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

// SourceResult holds the results of searching on a particular source for a
// particular query.
type SearchResult struct {
	Name  string
	Shows []Show
	Error error
}

// UpdateSeasonsAndEpisodes should be called to update a Show after, for
// example, deserialization from disk.
func UpdateSeasonsAndEpisodes(show *store.Show) error {
	var seasons []Season
	var err error

	if show.SourceName == "trakt" {
		seasons, err = seasonsWithTrakt(show)
	}
	if show.SourceName == "tvrage" {
		seasons, err = seasonsWithTvRage(show)
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

	return err
}

func seasonsWithTrakt(show *store.Show) ([]Season, error) {
	var seasons []Season

	client := traktClient()
	traktSeasons, result := client.Seasons().All(show.ID)
	if result.Err != nil {
		return seasons, result.Err
	}

	for i := 0; i < len(traktSeasons); i++ {
		season := Season{
			Season: traktSeasons[i].Number,
		}
		episodes, result := client.Episodes().AllBySeason(show.ID, traktSeasons[i].Number)
		if result.Err != nil {
			return seasons, result.Err
		}
		for _, episode := range episodes {
			if episode.FirstAired == nil {
				episode.FirstAired = &time.Time{}
			}
			season.Episodes = append(
				season.Episodes,
				Episode{
					Title:   episode.Title,
					AirDate: *episode.FirstAired,
					Episode: episode.Number,
				},
			)
		}
		seasons = append(seasons, season)
	}
	return seasons, nil
}

func seasonsWithTvRage(show *store.Show) ([]Season, error) {
	return TvRage{}.AllSeasonsAndEpisodes(show.ID)
}

// Search is the important function of this package. Call this to turn a user
// search string into a list of matches (which might be TV shows or movies).
func Search(q string) []SearchResult {
	var searchResults []SearchResult
	r := searchTrakt(q)
	searchResults = append(searchResults, r)
	r = searchTvRage(q)
	searchResults = append(searchResults, r)

	return searchResults
}

func traktClient() *trakt.Client {
	return trakt.NewClient(
		"01045164ed603042b53acf841b590f0e7b728dbff319c8d128f8649e2427cbe9",
		trakt.TokenAuth{AccessToken: "3b6f5bdba2fa56b086712d5f3f15b4e967f99ab049a6d3a4c2e56dc9c3c90462"},
	)
}

func searchTrakt(q string) SearchResult {
	searchResult := SearchResult{
		Name: "trakt",
	}

	results, response := traktClient().Shows().Search(q)
	if response.Err != nil {
		searchResult.Error = response.Err
		return searchResult
	}

	for _, result := range results {
		searchResult.Shows = append(
			searchResult.Shows,
			Show{Source: searchResult.Name, Title: result.Show.Title, ID: result.Show.IDs.Trakt},
		)
	}
	return searchResult
}

func searchTvRage(q string) SearchResult {
	searchResult := SearchResult{
		Name: "tvrage",
	}

	searchResult.Shows, searchResult.Error = TvRage{}.Search(q)
	return searchResult
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
