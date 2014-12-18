package store

import "github.com/haarts/getme/sources"

type Show struct {
	Title string
}

type Store struct {
	shows map[string]Show
}

// TODO deserialize from a bunch of files.
func Open() *Store {
	return &Store{
		shows: make(map[string]Show),
	}
}

// TODO adds serialization to a bunch of JSON files.
// Plan: each show is a dir in shows/. Each seasons is a dir in that. And each
// episode is a file in that. When an episode has been found and downloaded
// just rename the file. The file contains some meta data.

func (s *Store) CreateShow(m sources.Match) *Show {
	show := Show{m.Title}
	s.shows[m.Title] = show
	return &show
}
