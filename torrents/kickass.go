// TODO move this to it's own package. So we can have more of them w/o
// namespace clashes.
// TODO when moving it make sure that we're able to expose the logging system
// to the search engines.
package torrents

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
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

func (k Kickass) Search(show *store.Show) ([]Torrent, error) {
	seasonQueries := queriesForSeasons(show)
	seasonTorrents, err := k.torrentsForSeasons(seasonQueries)
	if err != nil {
		return nil, err
	}

	episodeQueries := queriesForEpisodes(show)
	episodeTorrents, err := k.torrentsForEpisodes(episodeQueries)
	if err != nil {
		return nil, err
	}

	return append(seasonTorrents, episodeTorrents...), err
}

func (k Kickass) torrentsForEpisodes(queries map[*store.Episode]string) ([]Torrent, error) {
	var torrents []Torrent

	for episode, query := range queries {
		results, err := k.runQuery(query)
		if err != nil {
			return nil, err
		}

		if len(results) == 0 {
			continue
		}

		bestTorrent := selectBest(results)

		log.WithFields(
			log.Fields{
				"query":       query,
				"bestTorrent": bestTorrent.OriginalName,
			}).Debug("query with best result")

		bestTorrent.AssociatedMedia = episode
		snippet.Score = bestTorrent.seeds
		show.StoreEpisodeSnippet(snippet)
		torrents = append(torrents, *bestTorrent)
	}

	return torrents, nil
}

func (k Kickass) torrentsForSeasons(queries map[*store.Season]string) ([]Torrent, error) {
	var torrents []Torrent

	for season, query := range queries {
		results, err := k.runQuery(query)
		if err != nil {
			return nil, err
		}

		var rejectNonSeason = func(ts []Torrent) []Torrent {
			var rs []Torrent
			for _, t := range ts {
				if strings.Contains(strings.ToLower(t.OriginalName), "season") {
					rs = append(rs, t)
				}
			}
			return rs
		}

		results = rejectNonSeason(results)
		if len(results) == 0 {
			continue
		}

		bestTorrent := selectBest(results)

		log.WithFields(
			log.Fields{
				"query":       query,
				"bestTorrent": bestTorrent.OriginalName,
			}).Debug("query with best result")

		bestTorrent.AssociatedMedia = season
		snippet.Score = bestTorrent.seeds
		show.StoreSeasonSnippet(snippet)
		torrents = append(torrents, *bestTorrent)
	}

	return torrents, nil
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

	sort.Sort(bySeeds(searchItems))

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

type bySeeds []kickassItem

func (a bySeeds) Len() int           { return len(a) }
func (a bySeeds) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a bySeeds) Less(i, j int) bool { return a[i].Seeds > a[j].Seeds }
