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

// TODO eye watering duplication ...
func torrentsForSeasons(show *sources.Show) ([]Torrent, error) {
	var torrents []Torrent

	seasons := show.PendingSeasons()

	for _, season := range seasons {
		bestSeasonSnippet := max(show.QuerySnippets.ForSeason)

		if bestSeasonSnippet.Score == 0 {
			torrent := searchForSeasonWithAlternative(season, show)
			if torrent != (Torrent{}) {
				torrents = append(torrents, torrent)
			}
		} else {
			// TODO handle err
			seasonTorrents, _ := searchKickass(
				seasonQueryAlternatives[bestSeasonSnippet.FormatSnippet](
					bestSeasonSnippet.TitleSnippet,
					season,
				),
			)
			if len(seasonTorrents) > 0 {
				torrent := selectBest(seasonTorrents)
				torrent.AssociatedMedia = season
				torrents = append(torrents, torrent)
			} else {
				torrent := searchForSeasonWithAlternative(season, show)
				if torrent != (Torrent{}) {
					torrent.AssociatedMedia = season
					torrents = append(torrents, torrent)
				}
			}
		}
	}

	return torrents, nil
}

func selectBest(torrents []Torrent) Torrent {
	return torrents[0] //most peers
}

// TODO eye watering duplication ...
func searchForEpisodeWithAlternative(episode *sources.Episode, show *sources.Show) Torrent {
	type alternative struct {
		torrent Torrent
		snippet sources.Snippet
	}
	var alternatives []alternative

	for k, alt := range episodeQueryAlternatives {
		for _, morpher := range titleMorpher {
			altTitle := morpher(show.Title)

			altTorrents, _ := searchKickass(
				alt(
					altTitle,
					episode,
				),
			)
			if len(altTorrents) > 0 {
				torrent := selectBest(altTorrents)
				if torrent != (Torrent{}) { // Arrrr surely there is something better
					torrent.AssociatedMedia = episode
					alternatives = append(
						alternatives,
						alternative{torrent, sources.Snippet{torrent.seeds, altTitle, k}},
					)
				}
			}
		}
	}

	var best alternative
	for _, alt := range alternatives {
		if alt.torrent.seeds > best.torrent.seeds {
			best = alt
		}
	}

	show.QuerySnippets.ForSeason = append(show.QuerySnippets.ForEpisode, best.snippet)
	return best.torrent
}

// TODO eye watering duplication ...
func searchForSeasonWithAlternative(season *sources.Season, show *sources.Show) Torrent {
	type alternative struct {
		torrent Torrent
		snippet sources.Snippet
	}
	var alternatives []alternative

	for k, alt := range seasonQueryAlternatives {
		for _, morpher := range titleMorpher {
			altTitle := morpher(show.Title)

			// Weird coupling between formatting string and argument order...
			altTorrents, _ := searchKickass(alt(altTitle, season))
			fmt.Printf("alt %+v\n", k)
			fmt.Printf("altTitle %+v\n", altTitle)
			if len(altTorrents) > 0 {
				torrent := selectBest(altTorrents)
				if torrent != (Torrent{}) { // Arrrr surely there is something better
					torrent.AssociatedMedia = season
					alternatives = append(
						alternatives,
						alternative{torrent, sources.Snippet{torrent.seeds, altTitle, k}},
					)
				}
			}
		}
	}

	var best alternative
	for _, alt := range alternatives {
		if alt.torrent.seeds > best.torrent.seeds {
			best = alt
		}
	}

	show.QuerySnippets.ForSeason = append(show.QuerySnippets.ForSeason, best.snippet)
	return best.torrent
}

// TODO eye watering duplication ...
func torrentsForEpisodes(show *sources.Show) ([]Torrent, error) {
	var torrents []Torrent

	episodes := show.PendingEpisodes()

	for _, episode := range episodes {
		bestEpisodeSnippet := max(show.QuerySnippets.ForEpisode)

		if bestEpisodeSnippet.Score == 0 {
			torrent := searchForEpisodeWithAlternative(episode, show)
			if torrent != (Torrent{}) {
				torrents = append(torrents, torrent)
			}
		} else {

			// TODO handle err
			episodeTorrents, _ := searchKickass(
				episodeQueryAlternatives[bestEpisodeSnippet.FormatSnippet](
					bestEpisodeSnippet.TitleSnippet,
					episode,
				),
			)
			if len(episodeTorrents) > 0 {
				torrent := selectBest(episodeTorrents)
				torrent.AssociatedMedia = episode
				torrents = append(torrents, torrent)
			} else {
				torrent := searchForEpisodeWithAlternative(episode, show)
				if torrent != (Torrent{}) {
					torrent.AssociatedMedia = episode
					torrents = append(torrents, torrent)
				}
			}
		}
	}

	return torrents, nil
}

func max(snippets []sources.Snippet) sources.Snippet {
	var bestSnippet sources.Snippet
	for _, snippet := range snippets {
		if snippet.Score >= bestSnippet.Score {
			bestSnippet = snippet
		}
	}
	return bestSnippet
}

func (i Item) torrentURL() string {
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

func isEnglish(i Item) bool {
	// Too weak a check but it is the easiest. I hope there aren't any series
	// with 'french' in the title.
	if strings.Contains(strings.ToLower(i.FileName), "french") {
		return false
	}
	return true
}
