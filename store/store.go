package store

import "github.com/haarts/getme/sources"

type Show struct {
	Title string
}

type Store struct {
	shows map[string]Show
}

func Open() *Store {
	return &Store{
		shows: make(map[string]Show),
	}
}

func (s *Store) CreateShow(m sources.Match) *Show {
	show := Show{m.Title}
	s.shows[m.Title] = show
	return &show
}
