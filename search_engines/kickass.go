package search_engines

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

type TorrentURL string

type Item struct {
	Title    string `xml:"title"`
	InfoHash string `xml:"infoHash"`
	Seeds    int    `xml:"seeds"`
	Peers    int    `xml:"peers"`
}

func (i Item) torrentURL() TorrentURL {
	return TorrentURL(fmt.Sprintf(torCacheURL, i.InfoHash))
}

type kickassResult struct {
	Channel struct {
		Items []Item `xml:"item"`
	} `xml:"channel"`
}

type BySeeds []Item

func (a BySeeds) Len() int           { return len(a) }
func (a BySeeds) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySeeds) Less(i, j int) bool { return a[i].Seeds > a[j].Seeds }

var kickassSearchURL = "https://kickass.so"
var torCacheURL = "https://torcache.net/torrent/%s.torrent"

func constructURL(episode string) string { //NOTE this url concat is broken but it's for tests...
	return fmt.Sprintf(kickassSearchURL+"/usearch/%s/?rss=1", url.QueryEscape(episode))
}

func Search(episodes []*sources.Episode) ([]TorrentURL, error) {
	var results []TorrentURL
	var episodeResult []kickassResult
	// TODO dont loop if a list of episodes span a complete season. Search for the season instead.
	for _, e := range episodes {
		queries := e.QueryNames()
		for _, q := range queries {
			body, err := getBody(q)
			if err != nil {
				return nil, err
			}

			var kr kickassResult
			err = xml.Unmarshal(body, &kr)
			if err != nil {
				return nil, err
			}
			episodeResult = append(episodeResult, kr)
		}
		var xs []Item
		for _, e := range episodeResult {
			xs = append(xs, e.Channel.Items...)
		}

		onlyEnglish := xs[:0]
		for _, x := range xs {
			if isEnglish(x, *e) {
				onlyEnglish = append(onlyEnglish, x)
			}
		}

		sort.Sort(BySeeds(onlyEnglish))

		best := getBest(onlyEnglish)
		results = append(results, best.torrentURL())
	}
	return results, nil
}

// TODO pick 1080p if no there pick 720p
func getBest(xs []Item) Item {
	return xs[0]
}

// English is when the title doesn't meantion any names
// Except when in combination with Sub(s)
// Except when the show name has a language in it
// Impl:
// get all languages from title
// if language present check if is in showname
// if language present check if title contains 'Subs'
// Impl: getting language from title
// create word boundaries
// search for country codes as words
// search for english name for language
// search for native name for language
func isEnglish(i Item, e sources.Episode) bool {
	// Too weak a check but it is the easiest
	if strings.Contains(strings.ToLower(e.ShowName()), "french") {
		return true
	}

	// Too weak a check but the vast majority of foreign languages seem french
	if strings.Contains(strings.ToLower(i.Title), "french") {
		return false
	}
	return true

}

func getBody(query string) ([]byte, error) {
	resp, err := http.Get(constructURL(query))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Search return non 200 status code: %d", resp.StatusCode))
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
