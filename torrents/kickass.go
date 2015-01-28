package torrents

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
)

type Kickass struct{}

const kickassName = "kickass"
const batchSize = 50

func init() {
	Register(kickassName, Kickass{})
}

type SearchResult struct {
	Title    string `xml:"title"`
	InfoHash string `xml:"infoHash"`
	Seeds    int    `xml:"seeds"`
	Peers    int    `xml:"peers"`
	FileName string `xml:"fileName"`
}

type kickassSearchResult struct {
	Channel struct {
		Items []SearchResult `xml:"item"`
	} `xml:"channel"`
}

type BySeeds []SearchResult

func (a BySeeds) Len() int           { return len(a) }
func (a BySeeds) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySeeds) Less(i, j int) bool { return a[i].Seeds > a[j].Seeds }

var kickassURL = "https://kickass.so"
var torCacheURL = "http://torcache.gs/torrent/%s.torrent"

// TODO Also review the code with https://github.com/golang/go/wiki/CodeReviewComments

func (k Kickass) Search(show *store.Show) ([]Torrent, error) {
	seasonTorrents, err := torrentsForSeasons(show)
	if err != nil {
		return nil, err
	}

	episodeTorrents, err := torrentsForEpisodes(show)
	if err != nil {
		return nil, err
	}

	return append(seasonTorrents, episodeTorrents...), err
}

func selectBest(torrents []Torrent) *Torrent {
	return &(torrents[0]) //most peers
}

func torrentsForEpisodes(show *store.Show) ([]Torrent, error) {
	var torrents []Torrent

	episodes := show.PendingEpisodes()
	sort.Sort(store.ByAirDate(episodes))
	min := math.Min(float64(len(episodes)), float64(batchSize))

	for _, s := range episodes[0:int(min)] {
		bestSnippet := show.BestEpisodeSnippet()
		var as []alt
		if _, ok := episodeQueryAlternatives[bestSnippet.FormatSnippet]; ok {
			as = append(as, alt{snippet: *bestSnippet})
		} else {
			for k, _ := range episodeQueryAlternatives {
				for _, morpher := range titleMorphers {
					as = addIfNew(as, morpher(show.Title), k)
				}
			}
		}

		for i := 0; i < len(as); i++ {
			results, _ := searchKickass(
				episodeQueryAlternatives[as[i].snippet.FormatSnippet](as[i].snippet.TitleSnippet, s),
			)
			if len(results) != 0 {
				as[i].torrent = selectBest(results)
			}
		}

		best := bestAlt(as)
		if best != nil {
			best.torrent.AssociatedMedia = s
			best.snippet.Score = best.torrent.seeds
			show.StoreEpisodeSnippet(best.snippet)
			torrents = append(torrents, *best.torrent)
		}
	}

	return torrents, nil
}

func addIfNew(as []alt, title, format string) []alt {
	newAlt := alt{
		snippet: store.Snippet{
			Score:         0,
			TitleSnippet:  title,
			FormatSnippet: format,
		},
	}
	for _, existing := range as {
		if newAlt.snippet.TitleSnippet == existing.snippet.TitleSnippet &&
			newAlt.snippet.FormatSnippet == existing.snippet.FormatSnippet {
			return as
		}
	}

	return append(as, newAlt)
}

func torrentsForSeasons(show *store.Show) ([]Torrent, error) {
	var torrents []Torrent

	for _, s := range show.PendingSeasons() {
		bestSnippet := show.BestSeasonSnippet()
		var as []alt
		if _, ok := seasonQueryAlternatives[bestSnippet.FormatSnippet]; ok {
			as = append(as, alt{snippet: *bestSnippet})
		} else {
			for k, _ := range seasonQueryAlternatives {
				for _, morpher := range titleMorphers {
					as = addIfNew(as, morpher(show.Title), k)
				}
			}
		}

		for i := 0; i < len(as); i++ {
			results, _ := searchKickass(
				seasonQueryAlternatives[as[i].snippet.FormatSnippet](as[i].snippet.TitleSnippet, s),
			)
			if len(results) != 0 {
				as[i].torrent = selectBest(results)
			}
		}

		best := bestAlt(as)
		if best != nil {
			best.torrent.AssociatedMedia = s
			best.snippet.Score = best.torrent.seeds
			show.StoreSeasonSnippet(best.snippet)
			torrents = append(torrents, *best.torrent)
		}
	}

	return torrents, nil
}

type alt struct {
	torrent *Torrent
	snippet store.Snippet
}

func bestAlt(as []alt) *alt {
	if len(as) == 0 {
		return nil
	}

	withTorrents := as[:0]
	for _, x := range as {
		if x.torrent != nil {
			withTorrents = append(withTorrents, x)
		}
	}

	if len(withTorrents) == 0 {
		return nil
	}

	best := withTorrents[0]
	for _, a := range withTorrents {
		if a.torrent.seeds > best.torrent.seeds {
			best = a
		}
	}
	return &best
}

func (i SearchResult) torrentURL() string {
	return fmt.Sprintf(torCacheURL, i.InfoHash)
}

func constructSearchURL(episode string) string {
	return fmt.Sprintf(kickassURL+"/usearch/%s/?rss=1", url.QueryEscape(episode))
}

func (k Kickass) request(URL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func searchKickass(query string) ([]Torrent, error) {
	req, err := (Kickass{}).request(constructSearchURL(query))
	if err != nil {
		return nil, err
	}

	var result kickassSearchResult

	err = sources.GetXML(req, &result)
	if err != nil {
		return nil, err
	}

	searchItems := result.Channel.Items

	// If we're going to reject torrents, we should do it here. (non english, whatever)
	// ...
	onlyEnglish := searchItems[:0]

	for _, x := range searchItems {
		if isEnglish(x) {
			onlyEnglish = append(onlyEnglish, x)
		}
	}

	sort.Sort(BySeeds(searchItems))

	var torrents []Torrent
	for _, searchItem := range searchItems {
		torrent := Torrent{searchItem.torrentURL(), searchItem.FileName, searchItem.Seeds, nil}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

func isEnglish(i SearchResult) bool {
	lowerCaseFileName := strings.ToLower(i.FileName)
	// Too weak a check but it is the easiest. I hope there aren't any series
	// with 'french' in the title.
	if strings.Contains(lowerCaseFileName, "french") {
		return false
	}

	if strings.Contains(lowerCaseFileName, "vostfr") {
		return false
	}

	regex := regexp.MustCompile(`\bITA\b`)
	if regex.MatchString(i.FileName) {
		return false
	}

	return true
}
