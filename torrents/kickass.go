// TODO move this to it's own package. So we can have more of them w/o
// namespace clashes.
// TODO when moving it make sure that we're able to expose the logging system
// to the search engines.
package torrents

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/haarts/getme/sources"
)

// TODO Also review the code with https://github.com/golang/go/wiki/CodeReviewComments

type Kickass struct {
	URL         string
	torCacheURL string
}

func NewKickass() *Kickass {
	return &Kickass{
		URL:         "https://kickass.to",
		torCacheURL: "http://torcache.gs/torrent/%s.torrent",
	}
}

func (k Kickass) Name() string {
	return "kickass"
}

func (k Kickass) Search(query string) ([]Torrent, error) {
	req, err := http.NewRequest("GET", k.constructSearchURL(query), nil)
	if err != nil {
		return nil, err
	}

	var result kickassSearchResult

	err = sources.GetXML(req, &result)
	if err != nil {
		return nil, err
	}

	searchItems := result.Channel.Items

	var torrents []Torrent
	for _, searchItem := range searchItems {
		url, err := searchItem.torrentURL(k.torCacheURL)
		if err != nil {
			return nil, err
		}
		torrent := Torrent{
			URL:      url,
			Filename: searchItem.FileName,
			Title:    searchItem.Title,
			seeds:    searchItem.Seeds,
		}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

func (k Kickass) constructSearchURL(episode string) string {
	return fmt.Sprintf(k.URL+"/usearch/%s/?rss=1", url.QueryEscape(episode))
}

type kickassSearchResult struct {
	Channel struct {
		Items []kickassItem `xml:"item"`
	} `xml:"channel"`
}

type kickassItem struct {
	Title    string `xml:"title"`
	InfoHash string `xml:"infoHash"`
	Seeds    int    `xml:"seeds"`
	Peers    int    `xml:"peers"`
	FileName string `xml:"fileName"`
}

func (i kickassItem) torrentURL(torCacheURL string) (*url.URL, error) {
	return url.Parse(fmt.Sprintf(torCacheURL, i.InfoHash))
}
