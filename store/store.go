package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/haarts/getme/sources"
)

// Store is the main access point for everything storage related.
type Store struct {
	shows    map[string]*sources.Show
	movies   map[string]*sources.Movie
	stateDir string
}

// Open gets the serialized data from disk and reconstitutes them.
func Open(stateDir string) (*Store, error) {
	err := ensureStateDir(stateDir)
	if err != nil {
		return nil, err
	}
	store := &Store{
		shows:    make(map[string]*sources.Show),
		stateDir: stateDir,
	}

	store.deserializeShows()
	//for k, v := range store.shows {
	//fmt.Printf("k %+v\n", k)
	//fmt.Printf("v %+v\n", v)
	//fmt.Printf("v.Seasons[0] %+v\n", v.Seasons[0])
	//}

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
func (s *Store) CreateShow(m *sources.Show) error {
	if _, ok := s.shows[m.Title]; ok {
		return fmt.Errorf("Show %s already exists.\n", m.Title)
	}

	s.shows[m.Title] = m

	return nil
}

// Shows returns a list of shows.
func (s *Store) Shows() map[string]*sources.Show {
	return s.shows
}

// Movies returns a list of movies. Currently nothing will every be returned
// since it's impossible to store anything.
func (s *Store) Movies() map[string]*sources.Movie {
	return s.movies
}

//TODO handle err
// TODO add another pass after deserialization to set the pointers to seasons and show right.
func (s *Store) deserializeShows() {
	files, err := ioutil.ReadDir(path.Join(s.stateDir, "shows"))
	if err != nil {
		fmt.Printf("err %+v\n", err)
	}

	for _, f := range files {
		matched, err := regexp.MatchString(".*.json", f.Name())
		if err != nil {
			fmt.Printf("err %+v\n", err)
		}
		if !matched {
			continue
		}

		var show sources.Show
		d, err := ioutil.ReadFile(path.Join(s.stateDir, "shows", f.Name()))
		if err != nil {
			fmt.Printf("err %+v\n", err)
		}
		err = json.Unmarshal(d, &show)
		if err != nil {
			fmt.Printf("err %+v\n", err)
		}

		s.shows[show.Title] = &show
	}
}

func ensureStateDir(stateDir string) error {
	dirs := []string{
		stateDir,
		path.Join(stateDir, "shows"),
		path.Join(stateDir, "movies"),
	}

	for _, d := range dirs {
		err := os.Mkdir(d, 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}

	return nil
}

func (s Store) store(show *sources.Show) error {
	b, err := json.MarshalIndent(show, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(s.stateDir, "shows", titleAsFileName(show.Title)+".json"), b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func titleAsFileName(title string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	return string(re.ReplaceAll([]byte(title), []byte("_")))
}
