// Package torrents provides the ability to search for torrents given a
// list of required items.
package torrents

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/store"
)

// batchSize controls how many torrent will be fetch in 1 go. This is
// important when downloading very long running series.
const batchSize = 50

var searchTimeout = 3 * time.Second

// Mark a piece of media as done. Currently only Show.
type Doner interface {
	Done()
}

type Torrent struct {
	URL             *url.URL
	Filename        string
	Title           string
	seeds           int
	AssociatedMedia Doner
}

type SearchEngine interface {
	Search(string) ([]Torrent, error)
	Name() string
}

var searchEngines = map[string]SearchEngine{
	"kickass":      NewKickass(),
	"torrentcd":    NewTorrentCD(),
	"extratorrent": ExtraTorrent{},
}

type queryJob struct {
	media   Doner
	snippet store.Snippet
	query   string
	season  int // to distinguish between episode and season jobs.
}

func Search(show *store.Show) ([]Torrent, error) {
	// torrents holds the torrents to complete a serie
	var torrents []Torrent

	// TODO perhaps mashing season and episode jobs together is a bad idea
	queryJobs := createQueryJobs(show)
	for _, queryJob := range queryJobs {
		torrent, err := executeJob(queryJob)
		if err != nil {
			continue
		}

		torrent.AssociatedMedia = queryJob.media
		queryJob.snippet.Score = torrent.seeds
		// *ouch* this type switch is ugly
		switch queryJob.media.(type) {
		case *store.Season:
			show.StoreSeasonSnippet(queryJob.snippet)
		case *store.Episode:
			show.StoreEpisodeSnippet(queryJob.snippet)
		default:
			panic("unknown media type")
		}
		torrents = append(torrents, *torrent)
	}

	return torrents, nil
}

func executeJob(job queryJob) (*Torrent, error) {
	results := searchWithFilters(job, isEnglish, isSeason)

	torrents := collectResultsWithTimeout(results)

	if len(torrents) == 0 {
		return nil, fmt.Errorf("No torrents found for %s", job.query)
	}

	sort.Sort(bySeeds(torrents))
	bestTorrent := torrents[0]

	log.WithFields(log.Fields{
		"torrent_url": bestTorrent.URL,
		"title":       bestTorrent.Title,
		"score":       bestTorrent.seeds,
	}).Info("Selected best torrent")

	return &bestTorrent, nil
}

func searchWithFilters(job queryJob, filters ...filter) chan []Torrent {
	// c emits the torrents found for one search request on one search engine
	c := make(chan []Torrent)
	for _, searchEngine := range searchEngines {
		go func(s SearchEngine) {
			torrents, err := s.Search(job.query)
			if err != nil {
				log.WithFields(log.Fields{
					"err":           err,
					"search_engine": s.Name(),
				}).Error("Search engine returned error")
			}
			torrents = applyFilters(job, torrents, isEnglish, isSeason)
			c <- torrents
		}(searchEngine)
	}

	return c
}

func collectResultsWithTimeout(results chan []Torrent) []Torrent {
	var torrentsFromAllEngines []Torrent
	timeout := time.After(searchTimeout)
	for i := 0; i < len(searchEngines); i++ {
		select {
		case result := <-results:
			torrentsFromAllEngines = append(torrentsFromAllEngines, result...)
		case <-timeout:
			log.Error("Search timed out")
		}
	}

	return torrentsFromAllEngines
}

func createQueryJobs(show *store.Show) []queryJob {
	seasonQueries := queriesForSeasons(show)
	episodeQueries := queriesForEpisodes(show)
	return append(seasonQueries, episodeQueries...)
}

func queriesForEpisodes(show *store.Show) []queryJob {
	episodes := show.PendingEpisodes()
	sort.Sort(store.ByAirDate(episodes))
	min := math.Min(float64(len(episodes)), float64(batchSize))

	queries := []queryJob{}
	for _, episode := range episodes[0:int(min)] {
		snippet := selectEpisodeSnippet(show)

		query := episodeQueryAlternatives[snippet.FormatSnippet](snippet.TitleSnippet, episode)
		queries = append(queries, queryJob{snippet: snippet, query: query, media: episode})
	}
	return queries
}

func queriesForSeasons(show *store.Show) []queryJob {
	queries := []queryJob{}
	for _, season := range show.PendingSeasons() {
		// ignore Season 0, which are specials and are rarely found and/or
		// interesting.
		if season.Season == 0 {
			continue
		}

		snippet := selectSeasonSnippet(show)

		query := seasonQueryAlternatives[snippet.FormatSnippet](snippet.TitleSnippet, season)
		queries = append(queries, queryJob{
			snippet: snippet,
			query:   query,
			media:   season,
			season:  season.Season,
		})
	}
	return queries
}

type filter func(queryJob, string) bool

func applyFilters(job queryJob, torrents []Torrent, filters ...filter) []Torrent {
	ok := []Torrent{}
	for _, torrent := range torrents {
		allGood := true
		for _, f := range filters {
			if allGood == false {
				break
			}
			allGood = f(job, torrent.Title)
		}
		if allGood {
			ok = append(ok, torrent)
		}
	}
	return ok
}

func isSeason(job queryJob, title string) bool {
	if job.season == 0 {
		return true
	}
	if strings.Contains(strings.ToLower(title), fmt.Sprintf("season %d", job.season)) {
		return true
	}
	return false
}

func isEnglish(_ queryJob, title string) bool {
	lowerCaseTitle := strings.ToLower(title)
	// Too weak a check but it is the easiest. I hope there aren't any series
	// with 'french' in the title.
	if strings.Contains(lowerCaseTitle, "french") {
		return false
	}

	if strings.Contains(lowerCaseTitle, "spanish") {
		return false
	}

	if strings.Contains(lowerCaseTitle, "español") {
		return false
	}

	// Ignore Version Originale Sous-Titrée en FRançais. Hard coded, French subtitles.
	if strings.Contains(lowerCaseTitle, "vostfr") {
		return false
	}

	// Ignore Italian (ITA) dubs.
	regex := regexp.MustCompile(`\bITA\b`)
	if regex.MatchString(title) {
		return false
	}

	// Ignore hard coded (HC) subtitles.
	regex = regexp.MustCompile(`\bHC\b`)
	if regex.MatchString(title) {
		return false
	}

	return true
}

func selectBest(torrents []Torrent) *Torrent {
	return &(torrents[0]) //most peers
}

type bySeeds []Torrent

func (a bySeeds) Len() int           { return len(a) }
func (a bySeeds) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bySeeds) Less(i, j int) bool { return a[i].seeds > a[j].seeds }
