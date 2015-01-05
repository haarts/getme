// Package search_engines provides the ability to search for torrents given a
// list of required items.
package search_engines

import "github.com/haarts/getme/sources"

type SearchEngine interface {
	Search([]sources.PendingItem) ([]Torrent, error)
}

var searchEngines = make(map[string]SearchEngine)

func Register(name string, searchEngine SearchEngine) {
	if _, dup := searchEngines[name]; dup {
		panic("search_engine: Register called twice for search engine " + name)
	}
	searchEngines[name] = searchEngine
}

// TODO this is only a staring point for pull torrents for the same search
// engines. I need to come up with a way to pick the best on duplciates.
func Search(items []sources.PendingItem) ([]Torrent, error) {
	var torrents []Torrent
	var lastError error
	for _, searchEngine := range searchEngines {
		ts, err := searchEngine.Search(items)
		torrents = append(torrents, ts...)
		lastError = err
	}
	return torrents, lastError
}
