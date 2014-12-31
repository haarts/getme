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
	stateDir string
}

// TODO deserialize from a bunch of files.
func Open(stateDir string) (*Store, error) {
	err := ensureStateDir(stateDir)
	if err != nil {
		return nil, err
	}

	return &Store{
		shows:    make(map[string]*sources.Show),
		stateDir: stateDir,
	}, nil
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

func ensureStateDir(stateDir string) error {
	dirs := []string{
		stateDir,
		path.Join(stateDir, "shows"),
		path.Join(stateDir, "movies"),
	}

	for _, d := range dirs {
		err := os.Mkdir(d, 0755)
		if err != nil {
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

	err = ioutil.WriteFile(path.Join(s.stateDir, "shows", titleAsFileName(show.Title)+".json"), b, 0755)
	if err != nil {
		return err
	}

	return nil
}

func titleAsFileName(title string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	return string(re.ReplaceAll([]byte(title), []byte("_")))
}

func (s *Store) CreateShow(m *sources.Show) error {
	if _, ok := s.shows[m.Title]; ok {
		return errors.New(fmt.Sprintf("Show %s already exists.\n", m.Title))
	}
	s.shows[m.Title] = m
	return nil
}
