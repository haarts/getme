// TODO move this to it's own package. So we can have more of them w/o
// namespace clashes.
// TODO when moving it make sure that we're able to expose the logging system
// to the search engines.
package torrents

import (
	"fmt"
	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"

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
	log.WithFields(log.Fields{
		"query": query,
	}).Debug("Querying Kickass")
	return k.runQuery(query)
}

func (k Kickass) constructSearchURL(episode string) string {
	return fmt.Sprintf(k.URL+"/usearch/%s/?rss=1", url.QueryEscape(episode))
}

func (k Kickass) request(URL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (k Kickass) runQuery(query string) ([]Torrent, error) {
	req, err := k.request(k.constructSearchURL(query))
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
		if isEnglish(x.FileName) {
			onlyEnglish = append(onlyEnglish, x)
		}
	}

	var torrents []Torrent
	for _, searchItem := range searchItems {
		torrent := Torrent{searchItem.torrentURL(k.torCacheURL), searchItem.FileName, searchItem.Seeds, nil}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
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

func (i kickassItem) torrentURL(torCacheURL string) string {
	return fmt.Sprintf(torCacheURL, i.InfoHash)
}
