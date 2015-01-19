// Package store handles the persistence of GetMe. Currently it's all stored as
// JSON on disk.
package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"regexp"

	"github.com/haarts/getme/config"
)

var conf = config.Config()
var log = config.Log()

// Store is the main access point for everything storage related.
type Store struct {
	shows  map[string]*Show
	movies map[string]*Movie
}

// Open gets the serialized data from disk and reconstitutes them.
func Open() (*Store, error) {
	store := &Store{
		shows: make(map[string]*Show),
	}

	store.deserializeShows()

	return store, nil
}

// Close writes the, in memory, store value to disk. Do NOT forget to call this
// if you want to persist your data!
func (s Store) Close() error {
	for _, show := range s.shows {
		err := s.store(show)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateShow adds a show to the store. It does NOT persist it to disk yet! Use
// Close for this.
func (s *Store) CreateShow(m *Show) error {
	if _, ok := s.shows[m.Title]; ok {
		return fmt.Errorf("Show %s already exists.\n", m.Title)
	}

	s.shows[m.Title] = m

	return nil
}

// Shows returns a list of shows.
func (s *Store) Shows() map[string]*Show {
	return s.shows
}

// Movies returns a list of movies. Currently nothing will every be returned
// since it's impossible to store anything.
func (s *Store) Movies() map[string]*Movie {
	return s.movies
}

// TODO probably return the error
func (s *Store) deserializeShows() {
	files, err := ioutil.ReadDir(path.Join(conf.StateDir, "shows"))
	if err != nil {
		fmt.Printf("err %+v\n", err) // TODO log.Error
	}

	for _, f := range files {
		matched, err := regexp.MatchString(".*.json", f.Name())
		if err != nil {
			fmt.Printf("err %+v\n", err) // TODO log.Error
		}
		if !matched {
			continue
		}

		var show Show
		d, err := ioutil.ReadFile(path.Join(conf.StateDir, "shows", f.Name()))
		if err != nil {
			fmt.Printf("err %+v\n", err) // TODO log.Error
		}
		err = json.Unmarshal(d, &show)
		if err != nil {
			fmt.Printf("err %+v\n", err) // TODO log.Error
		}

		s.shows[show.Title] = &show
	}
}

func (s Store) store(show *Show) error {
	b, err := json.MarshalIndent(show, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(conf.StateDir, "shows", titleAsFileName(show.Title)+".json"), b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func titleAsFileName(title string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	return string(re.ReplaceAll([]byte(title), []byte("_")))
}
