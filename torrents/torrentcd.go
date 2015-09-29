package torrents

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/haarts/getme/sources"
)

type TorrentCD struct {
	URL string
}

func NewTorrentCD() *TorrentCD {
	return &TorrentCD{
		URL: "http://torrent.cd",
	}
}

func (t TorrentCD) Name() string {
	return "torrentCD"
}

func (t TorrentCD) Search(query string) ([]Torrent, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(t.URL+"/torrents/xml?q=%s", url.QueryEscape(query)),
		nil,
	)

	var result torrentCDSearchResult
	err = sources.GetXML(req, &result)
	if err != nil {
		return nil, err
	}

	var torrents []Torrent
	for _, item := range result.Channel.Items {
		torrent := Torrent{
			URL:      item.torrentURL(),
			Title:    item.Title,
			Filename: item.Filename(),
			seeds:    item.Seed,
		}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}

type torrentCDSearchResult struct {
	Channel struct {
		Items []torrentCDItem `xml:"item"`
	} `xml:"channel"`
}

type torrentCDItem struct {
	Link  string `xml:"link"`
	Title string `xml:"title"`
	Seed  int    `xml:"seed"`
}

func (t torrentCDItem) torrentURL() string {
	return strings.Replace(t.Link, "http://torrent.cd/", "http://torrent.cd/torrents/download/", 1)
}

func (t torrentCDItem) Filename() string {
	url, _ := url.Parse(t.Link)
	return path.Base(url.Path)
}
