package torrents

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/sources"
)

type TorrentCD struct {
	URL string
}

func NewTorrentCD() *TorrentCD {
	return &TorrentCD{
		URL: "http://torrentcd.net",
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
	if err == io.EOF { // the result contain non XML when nothing is found
		log.WithFields(log.Fields{
			"search_engine": t.Name(),
			"url":           req.URL,
		}).Debug("No torrents found.")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var torrents []Torrent
	for _, item := range result.Channel.Items {
		url, err := item.torrentURL()
		if err != nil {
			return nil, err
		}

		torrent := Torrent{
			URL:      url,
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

func (t torrentCDItem) torrentURL() (*url.URL, error) {
	return url.Parse(strings.Replace(t.Link, "http://torrentcd.net/", "http://torrentcd.net/torrents/download/", 1))
}

func (t torrentCDItem) Filename() string {
	url, _ := url.Parse(t.Link)
	return path.Base(url.Path)
}
