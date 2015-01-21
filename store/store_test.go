package store_test

import (
	"os"
	"path"
	"testing"

	"github.com/haarts/getme/config"
	"github.com/haarts/getme/store"
)

func TestClose(t *testing.T) {
	testDir := "test_state_dir"
	os.MkdirAll(path.Join(testDir, "shows"), 0755)
	defer func() {
		os.RemoveAll(testDir)
	}()
	config.Config().StateDir = testDir

	s, _ := store.Open()
	show := store.Show{Title: "my show"}
	s.CreateShow(&show)

	s.Close()
	if _, err := os.Stat(path.Join(testDir, "shows", "my_show.json")); os.IsNotExist(err) {
		t.Error("Expected show to be stored as file.")
	}
}

func TestCreateDuplicateShow(t *testing.T) {
	testDir := "test_state_dir"
	os.MkdirAll(path.Join(testDir, "shows"), 0755)
	defer func() {
		os.RemoveAll(testDir)
	}()
	config.Config().StateDir = testDir

	s, _ := store.Open()

	show := store.Show{Title: "my show"}
	s.CreateShow(&show)
	err := s.CreateShow(&show)
	if err == nil {
		t.Error("Expected not to be able to store same shows.")
	}
}

func TestReadShows(t *testing.T) {
	testDir := "test_state_dir"
	os.MkdirAll(path.Join(testDir, "shows"), 0755)
	defer func() {
		os.RemoveAll(testDir)
	}()
	config.Config().StateDir = testDir

	os.Link(path.Join("testdata", "my_show.json"), path.Join(testDir, "shows", "my_show.json"))

	s, _ := store.Open()
	if len(s.Shows()) != 1 {
		t.Error("Expected to have read 1 show, got:", len(s.Shows()))
	}
	if _, ok := s.Shows()["my show"]; !ok {
		t.Error("Expected to find 'my show'.")
	}
}
