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

type Store struct {
	shows    map[string]*sources.Show
	stateDir string
}

// TODO deserialize from a bunch of files.
func Open(stateDir string) *Store {
	err := os.Mkdir(stateDir, 0755)
	if err != nil {
		fmt.Printf("err %+v\n", err) // TODO handle err properly
	}
	return &Store{
		shows:    make(map[string]*sources.Show),
		stateDir: stateDir,
	}
}

// TODO flush to disk
func (s Store) Close() {
	for _, show := range s.shows {
		s.store(show)
	}
}

func (s Store) store(show *sources.Show) {
	b, err := json.Marshal(show)
	if err != nil {
		fmt.Printf("err %+v\n", err) //TODO handle err properly
	}

	err = ioutil.WriteFile(path.Join(s.stateDir, titleAsFileName(show.Title)+".json"), b, 0755)
	if err != nil {
		fmt.Printf("err %+v\n", err) //TODO handle err properly
	}
}

func titleAsFileName(title string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	return string(re.ReplaceAll([]byte(title), []byte("_")))
}

// TODO adds serialization to a bunch of JSON files.
// Plan: each show is a dir in shows/. Each seasons is a dir in that. And each
// episode is a file in that. When an episode has been found and downloaded
// just rename the file. The file contains some meta data.

func (s *Store) CreateShow(m *sources.Show) {
	s.shows[m.Title] = m
}
