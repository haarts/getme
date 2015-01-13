package torrents

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/haarts/getme/sources"
)

type Kickass struct{}

const kickassName = "kickass"

func init() {
	Register(kickassName, Kickass{})
}

// TODO rename to SearchResult. Item is too generic.
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
var torCacheURL = "http://torcache.net/torrent/%s.torrent"

// TODO convert everything to func (s Show) Smt(season Season) {} eg value
// receiver single letter, arguments all out.
// TODO Also review the code with https://github.com/golang/go/wiki/CodeReviewComments

func (k Kickass) Search(show *sources.Show) ([]Torrent, error) {
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

func selectBest(torrents []Torrent) Torrent {
	return torrents[0] //most peers
}

func torrentsForEpisodes(show *sources.Show) ([]Torrent, error) {
	var torrents []Torrent

	for _, s := range show.PendingEpisodes() {
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

		best := alts(as).best()
		if best != nil {
			best.torrent.AssociatedMedia = s
			best.snippet.Score = best.torrent.seeds
			show.StoreEpisodeSnippet(best.snippet)
			torrents = append(torrents, best.torrent)
		}
	}

	return torrents, nil
}

func addIfNew(as []alt, title, format string) []alt {
	newAlt := alt{
		snippet: sources.Snippet{
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

func torrentsForSeasons(show *sources.Show) ([]Torrent, error) {
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

		best := alts(as).best()
		if best != nil {
			best.torrent.AssociatedMedia = s
			best.snippet.Score = best.torrent.seeds
			show.StoreSeasonSnippet(best.snippet)
			torrents = append(torrents, best.torrent)
		}
	}

	return torrents, nil
}

type alt struct {
	torrent Torrent
	snippet sources.Snippet
}

type alts []alt

func (as alts) best() *alt {
	var best alt
	for _, a := range as {
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

// TODO this could use the 'get' method in sources.go
func searchKickass(query string) ([]Torrent, error) {
	resp, err := http.Get(constructSearchURL(query))
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Search returned non 200 status code: %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result kickassSearchResult

	err = xml.Unmarshal(body, &result)
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
	// Too weak a check but it is the easiest. I hope there aren't any series
	// with 'french' in the title.
	if strings.Contains(strings.ToLower(i.FileName), "french") {
		return false
	}
	return true
}
