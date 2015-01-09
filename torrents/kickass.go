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
	Register(KICKASS, Kickass{})
}

type Torrent struct {
	URL          string
	PendingItem  sources.PendingItem
	OriginalName string
}

type Item struct {
	Title    string `xml:"title"`
	InfoHash string `xml:"infoHash"`
	Seeds    int    `xml:"seeds"`
	Peers    int    `xml:"peers"`
	FileName string `xml:"fileName"`
}

type kickassSearchResult struct {
	Channel struct {
		Items []Item `xml:"item"`
	} `xml:"channel"`
}

type BySeeds []Item

func (a BySeeds) Len() int           { return len(a) }
func (a BySeeds) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySeeds) Less(i, j int) bool { return a[i].Seeds > a[j].Seeds }

var kickassURL = "https://kickass.so"
var torCacheURL = "http://torcache.net/torrent/%s.torrent"

func (k Kickass) Search(items []sources.PendingItem) ([]Torrent, error) {
	var results []Torrent
	// TODO parallel execution
	for _, item := range items {
		best, err := getBestTorrentFor(item)
		if err != nil {
			return nil, err
		}
		if best == nil {
			continue
		}

		torrent := Torrent{best.torrentURL(), item, best.FileName}
		results = append(results, torrent)
	}
	return results, nil
}

func (i Item) torrentURL() string {
	return fmt.Sprintf(torCacheURL, i.InfoHash)
}

func constructSearchURL(episode string) string {
	return fmt.Sprintf(kickassURL+"/usearch/%s/?rss=1", url.QueryEscape(episode))
}

func getBestTorrentFor(e sources.PendingItem) (*Item, error) {
	var results []Item
	for _, q := range e.QueryNames {
		body, err := searchKickass(q)
		if err != nil { // No luck for this query.
			continue
		}

		var result kickassSearchResult
		err = xml.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		results = append(results, result.Channel.Items...)
	}

	if len(results) == 0 {
		// NOTE Just return nil, nil. Having no search results isn't an error.
		return nil, nil
	}

	onlyEnglish := results[:0]

	for _, x := range results {
		if isEnglish(x, e) {
			onlyEnglish = append(onlyEnglish, x)
		}
	}

	sort.Sort(BySeeds(onlyEnglish))
	best := getBest(onlyEnglish)
	return &best, nil
}

// TODO pick 1080p if no there pick 720p
func getBest(xs []Item) Item {
	return xs[0]
}

// TODO
// search for english name for foreign language
// search for native name for foreign language
// reject if found
func isEnglish(i Item, e sources.PendingItem) bool {
	// Too weak a check but it is the easiest
	if strings.Contains(strings.ToLower(e.ShowTitle), "french") {
		return true
	}

	// Too weak a check but the vast majority of foreign languages seem french
	if strings.Contains(strings.ToLower(i.Title), "french") {
		return false
	}
	return true

}

func searchKickass(query string) ([]byte, error) {
	resp, err := http.Get(constructSearchURL(query))
	defer resp.Body.Close()

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

	return body, nil
}
