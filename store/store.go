package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/haarts/getme/sources"
)

type Store struct {
	shows    map[string]*sources.Show
	movies   map[string]*sources.Movie
	stateDir string
}

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

	return store, nil
}

func (s Store) Close() error {
	for _, show := range s.shows {
		err := s.store(show)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) CreateShow(m *sources.Show) error {
	if _, ok := s.shows[m.Title]; ok {
		return errors.New(fmt.Sprintf("Show %s already exists.\n", m.Title))
	}

	s.shows[m.Title] = m

	return nil
}

func (s *Store) Shows() map[string]*sources.Show {
	return s.shows
}

func (s *Store) Movies() map[string]*sources.Movie {
	return s.movies
}

//TODO handle err
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
	b, err := json.Marshal(show)
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
