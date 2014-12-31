package store_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
)

func TestClose(t *testing.T) {
	testDir := "test_state_dir"
	defer func() {
		os.RemoveAll(testDir)
	}()

	s, _ := store.Open(testDir)
	show := sources.Show{Title: "my show"}
	s.CreateShow(&show)

	s.Close()
	if _, err := os.Stat(path.Join(testDir, "shows", "my_show.json")); os.IsNotExist(err) {
		t.Error("Expected show to be stored as file.")
	}
}

func TestCreateDuplicateShow(t *testing.T) {
	testDir := "test_state_dir"
	defer func() {
		os.RemoveAll(testDir)
	}()
	s, _ := store.Open(testDir)

	show := sources.Show{Title: "my show"}
	s.CreateShow(&show)
	err := s.CreateShow(&show)
	if err == nil {
		t.Error("Expected not to be able to store same shows.")
	}
}

func TestReadShows(t *testing.T) {
	testDir := "test_state_dir"
	defer func() {
		fmt.Printf("testDir %+v\n", testDir)
		os.RemoveAll(testDir)
	}()

	os.MkdirAll(path.Join(testDir, "shows"), 0755)
	os.Link(path.Join("fixtures", "my_show.json"), path.Join(testDir, "shows", "my_show.json"))

	s, _ := store.Open(testDir)
}
