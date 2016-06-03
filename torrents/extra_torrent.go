package torrents

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/sources"
)

type ExtraTorrent struct {
	URL string
}

type extraTorrentSearchResult struct {
	Channel struct {
		Items []extraTorrentItem `xml:"item"`
	} `xml:"channel"`
}

type extraTorrentItem struct {
	Title     string `xml:"title"`
	InfoHash  string `xml:"info_hash"`
	Seeders   string `xml:"seeders"`
	Enclosure struct {
		URL string `xml:"url,attr"`
	} `xml:"enclosure"`
}

func NewExtraTorrent() *ExtraTorrent {
	return &ExtraTorrent{
		URL: "https://extratorrent.cc",
	}
}

func (e ExtraTorrent) Name() string {
	return "extratorrent"
}

func (e ExtraTorrent) Search(query string) ([]Torrent, error) {
	req, err := http.NewRequest(
		"GET",
		e.URL+fmt.Sprintf("/search/?search=%s&new=1&x=0&y=0", query),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var result extraTorrentSearchResult

	err = sources.GetXML(req, &result)
	if er, ok := err.(sources.RequestError); ok {
		if er.ResponseCode == 404 {
			log.WithFields(log.Fields{
				"search_engine": e.Name(),
				"url":           req.URL,
			}).Debug("No torrents found.")
			return nil, nil
		}
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	var torrents []Torrent
	for _, item := range result.Channel.Items {
		url, err := url.Parse(item.Enclosure.URL)
		if err != nil {
			continue
		}

		if item.Seeders == "---" {
			item.Seeders = "0"
		}
		seeds, err := strconv.Atoi(item.Seeders)
		if err != nil {
			continue
		}

		parts := strings.Split(url.Path, "/")
		filename := parts[len(parts)-1]

		torrent := Torrent{
			URL:      url,
			Filename: filename,
			Title:    item.Title,
			seeds:    seeds,
		}
		torrents = append(torrents, torrent)
	}

	return torrents, nil
}
