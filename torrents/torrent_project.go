package torrents

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/haarts/getme/sources"

	log "github.com/Sirupsen/logrus"
)

type TorrentProject struct {
	URL         string
	TorCacheURL string
}

// this is one of the oddest returns I've seen
type torrentProjectSearchResult struct {
	One   torrentProjectItem `json:"1"`
	Two   torrentProjectItem `json:"2"`
	Three torrentProjectItem `json:"3"`
	Four  torrentProjectItem `json:"4"`
	Five  torrentProjectItem `json:"5"`
	Six   torrentProjectItem `json:"6"`
	Seven torrentProjectItem `json:"7"`
	Eight torrentProjectItem `json:"8"`
	Nine  torrentProjectItem `json:"9"`
	Ten   torrentProjectItem `json:"10"`
}

type torrentProjectItem struct {
	Title       string `json:"title"`
	Seeds       int    `json:"seeds"`
	TorrentHash string `json:"torrent_hash"`
}

func NewTorrentProject() *TorrentProject {
	return &TorrentProject{
		URL:         "https://torrentproject.se",
		TorCacheURL: "http://torcache.net/torrent/%s.torrent",
	}
}

func (t TorrentProject) Name() string {
	return "torrentproject"
}

func (t TorrentProject) Search(query string) ([]Torrent, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(t.URL+"/?s=%s&out=json&orderby=seeds", query),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var result torrentProjectSearchResult
	err = sources.GetJSON(req, &result)
	if e, ok := err.(sources.RequestError); ok {
		if e.ResponseCode == 404 {
			log.WithFields(log.Fields{
				"search_engine": t.Name(),
				"url":           req.URL,
			}).Debug("No torrents found.")
			return nil, nil
		}
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	convert := func(item torrentProjectItem) Torrent {
		url, err := item.torrentURL(t.TorCacheURL)
		if err != nil {
			log.WithFields(log.Fields{
				"err":     err,
				"torrent": item.Title,
			}).Error("failed to construct torrent url")
		}
		return Torrent{
			URL:      url,
			Filename: item.Title + ".torrent",
			Title:    item.Title,
			seeds:    item.Seeds,
		}
	}

	var torrents []Torrent
	torrents = append(torrents, convert(result.One))
	torrents = append(torrents, convert(result.Two))
	torrents = append(torrents, convert(result.Three))
	torrents = append(torrents, convert(result.Four))
	torrents = append(torrents, convert(result.Five))
	torrents = append(torrents, convert(result.Six))
	torrents = append(torrents, convert(result.Seven))
	torrents = append(torrents, convert(result.Eight))
	torrents = append(torrents, convert(result.Nine))
	torrents = append(torrents, convert(result.Ten))

	return torrents, nil
}

func (t torrentProjectItem) torrentURL(torCacheURL string) (*url.URL, error) {
	return url.Parse(fmt.Sprintf(torCacheURL, t.TorrentHash))
}
