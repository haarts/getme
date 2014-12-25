package store

import (
	"fmt"

	"github.com/haarts/getme/sources"
)

// TODO deprecated this. Just use sources.Show
type Show struct {
	Title string
}

type Store struct {
	shows    map[string]Show
	stateDir string
}

// TODO deserialize from a bunch of files.
func Open(stateDir string) *Store {
	return &Store{
		shows:    make(map[string]Show),
		stateDir: stateDir,
	}
}

// TODO flush to disk
func (s Store) Close() {
	for _, show := range s.shows {

		fmt.Printf("s %+v\n", show)
	}
}

// TODO adds serialization to a bunch of JSON files.
// Plan: each show is a dir in shows/. Each seasons is a dir in that. And each
// episode is a file in that. When an episode has been found and downloaded
// just rename the file. The file contains some meta data.

func (s *Store) CreateShow(m *sources.Show) *Show {
	show := Show{m.Title}
	s.shows[m.Title] = show
	return &show
}
