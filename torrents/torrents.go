// Package torrents provides the ability to search for torrents given a
// list of required items.
package torrents

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/haarts/getme/store"
)

// batchSize controls how many torrent will be fetch in 1 go. This is
// important when downloading very long running series.
const batchSize = 50

// Mark a piece of media as done. Currently only Show.
type Doner interface {
	Done()
}

type Torrent struct {
	URL             string
	OriginalName    string
	seeds           int
	AssociatedMedia Doner
}

type SearchEngine interface {
	Search(*store.Show) ([]Torrent, error)
}

var searchEngines = map[string]SearchEngine{
	"kickass":      Kickass{},
	"torrentcd":    TorrentCD{},
	"extratorrent": ExtraTorrent{},
}

type queryJob struct {
	media   Doner
	snippit store.Snippet
	query   string
}

// TODO this is only a starting point for pull torrents for the same search
// engines. I need to come up with a way to pick the best on duplciates.
func Search(show *store.Show) ([]Torrent, error) {
	var torrents []Torrent
	var lastError error
	for _, searchEngine := range searchEngines {
		ts, err := searchEngine.Search(show)
		torrents = append(torrents, ts...)
		lastError = err
	}
	return torrents, lastError
}

func Search(show *store.Show) ([]Torrent, error) {
	// torrents holds the torrents to complete a serie
	var torrents []Torrent
	queryJobs := createQueryJobs(show)
	for _, queryJob := range queryJobs {
		torrents = append(torrents, executeJob(queryJob))
	}

	return torrents
}

func executeJob(queryJob queryJob) Torrent {
	// c emits the torrents found for one search request on one search engine
	c := make(chan []Torrent)
	for _, searchEngine := range searchEngines {
		go func(s SearchEngine) {
			torrents, err := s.Search(queryJob.query)
			applyFilters(torrents, isEnglish, isSeason)
			c <- torrents
		}(searchEngine)
	}

	var torrentsPerQuery []Torrent
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(searchEngines); i++ {
		select {
		case result := <-c:
			torrents = append(torrentsPerQuery, result)
		case <-timeout:
			log.Error("Search timed out")
		}
	}
	// update score? // update snippet?
}

type filter func([]queryJob) []queryJob

func applyFilters(results []queryJob, filers ...filter) []queryJob {

}

func createQueryJobs(show *store.Show) []queryJob {
	seasonQueries := queriesForSeasons(show)
	episodeQueries := queriesForEpisodes(show)
	return append(seasonQueries, episodeQueries)
}

func queriesForEpisodes(show *store.Show) map[*store.Episode]string {
	episodes := show.PendingEpisodes()
	sort.Sort(store.ByAirDate(episodes))
	min := math.Min(float64(len(episodes)), float64(batchSize))

	queries := map[*store.Episode]string{}
	for _, episode := range episodes[0:int(min)] {
		snippet := selectEpisodeSnippet(show)

		query := episodeQueryAlternatives[snippet.FormatSnippet](snippet.TitleSnippet, episode)
		queries[episode] = query
	}
	return queries
}

func queriesForSeasons(show *store.Show) map[*store.Season]string {
	queries := map[*store.Season]string{}
	for _, season := range show.PendingSeasons() {
		// ignore Season 0, which are specials and are rarely found and/or
		// interesting.
		if season.Season == 0 {
			continue
		}

		snippet := selectSeasonSnippet(show)

		query := seasonQueryAlternatives[snippet.FormatSnippet](snippet.TitleSnippet, season)
		queries[season] = query
	}
	return queries
}

func isEnglish(fileName string) bool {
	lowerCaseFileName := strings.ToLower(fileName)
	// Too weak a check but it is the easiest. I hope there aren't any series
	// with 'french' in the title.
	if strings.Contains(lowerCaseFileName, "french") {
		return false
	}

	if strings.Contains(lowerCaseFileName, "spanish") {
		return false
	}

	if strings.Contains(lowerCaseFileName, "español") {
		return false
	}

	// Ignore Version Originale Sous-Titrée en FRançais. Hard coded, French subtitles.
	if strings.Contains(lowerCaseFileName, "vostfr") {
		return false
	}

	// Ignore Italian (ITA) dubs.
	regex := regexp.MustCompile(`\bITA\b`)
	if regex.MatchString(fileName) {
		return false
	}

	// Ignore hard coded (HC) subtitles.
	regex = regexp.MustCompile(`\bHC\b`)
	if regex.MatchString(fileName) {
		return false
	}

	return true
}

func selectBest(torrents []Torrent) *Torrent {
	return &(torrents[0]) //most peers
}
