package torrents

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/store"
)

var seasonQueryAlternatives = map[string]func(string, *store.Season) string{
	"%s season %d": func(title string, season *store.Season) string {
		return fmt.Sprintf("%s season %d", title, season.Season)
	},
}

var episodeQueryAlternatives = map[string]func(string, *store.Episode) string{
	"%s S%02dE%02d": func(title string, episode *store.Episode) string {
		return fmt.Sprintf("%s S%02dE%02d", title, episode.Season(), episode.Episode)
	},
	"%s %dx%d": func(title string, episode *store.Episode) string {
		return fmt.Sprintf("%s %dx%d", title, episode.Season(), episode.Episode)
	},
	// This is a bit of a gamble. I, now, no longer make the
	// distinction between a daily series and a regular one:
	"%s %d %02d %02d": func(title string, episode *store.Episode) string {
		y, m, d := episode.AirDate.Date()
		return fmt.Sprintf("%s %d %02d %02d", title, y, m, d)
	},
}

var titleMorphers = [...]func(string) string{
	func(title string) string { //noop
		return title
	},
	func(title string) string {
		re := regexp.MustCompile("[^ a-zA-Z0-9]")
		newTitle := string(re.ReplaceAll([]byte(title), []byte("")))
		return newTitle
	},
	func(title string) string {
		return truncateToNParts(title, 5)
	},
	func(title string) string {
		return truncateToNParts(title, 4)
	},
	func(title string) string {
		return truncateToNParts(title, 3)
	},
}

//type alt struct {
//torrent *Torrent
//snippet store.Snippet
//}

//func bestAlt(as []alt) *alt {
//if len(as) == 0 {
//return nil
//}

//withTorrents := as[:0]
//for _, x := range as {
//if x.torrent != nil {
//withTorrents = append(withTorrents, x)
//}
//}

//if len(withTorrents) == 0 {
//return nil
//}

//best := withTorrents[0]
//for _, a := range withTorrents {
//if a.torrent.seeds > best.torrent.seeds {
//best = a
//}
//}
//return &best
//}

func selectEpisodeSnippet(show *store.Show) store.Snippet {
	if len(show.QuerySnippets.ForEpisode) == 0 || isExplore() {
		// select random snippet
		var snippets []store.Snippet
		for k, _ := range episodeQueryAlternatives {
			for _, morpher := range titleMorphers {
				snippets = append(
					snippets,
					store.Snippet{
						Score:         0,
						TitleSnippet:  morpher(show.Title),
						FormatSnippet: k,
					},
				)
			}
		}
		snippet := snippets[rand.Intn(len(snippets))]
		log.WithFields(
			log.Fields{
				"title_snippet":  snippet.TitleSnippet,
				"format_snippet": snippet.FormatSnippet,
			}).Debug("Random snippet")
		return snippet
	}

	// select the current best
	return show.BestEpisodeSnippet()
}

func addIfNew(as []alt, title, format string) []alt {
	newAlt := alt{
		snippet: store.Snippet{
			Score:         0,
			TitleSnippet:  title,
			FormatSnippet: format,
		},
	}
	for _, existing := range as {
		if newAlt.snippet.TitleSnippet == existing.snippet.TitleSnippet &&
			newAlt.snippet.FormatSnippet == existing.snippet.FormatSnippet {
			return as
		}
	}

	return append(as, newAlt)
}

func selectSeasonSnippet(show *store.Show) store.Snippet {
	if len(show.QuerySnippets.ForSeason) == 0 || isExplore() {
		// select random snippet
		var snippets []store.Snippet
		for k, _ := range seasonQueryAlternatives {
			for _, morpher := range titleMorphers {
				snippets = append(
					snippets,
					store.Snippet{
						Score:         0,
						TitleSnippet:  morpher(show.Title),
						FormatSnippet: k,
					},
				)
			}
		}
		return snippets[rand.Intn(len(snippets))]
	}

	// select the current best
	return show.BestSeasonSnippet()
}

func isExplore() bool {
	if rand.Intn(10) == 0 { // explore
		return true
	}
	return false
}

func truncateToNParts(title string, n int) string {
	parts := strings.Split(title, " ")
	if len(parts) < n {
		return title
	}
	return strings.Join(parts[:n], " ")
}
