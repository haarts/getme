package store

import (
	"fmt"

	"github.com/haarts/getme/sources"
)

type Store struct {
	shows    map[string]*sources.Show
	stateDir string
}

// TODO deserialize from a bunch of files.
func Open(stateDir string) *Store {
	return &Store{
		shows:    make(map[string]*sources.Show),
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

func (s *Store) CreateShow(m *sources.Show) {
	s.shows[m.Title] = m
}
