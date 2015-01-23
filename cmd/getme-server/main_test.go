package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/haarts/getme/store"
)

func defaultSearchHandler() searchHandler {
	// Open our connection and setup our handler.
	testDir := "testdata/store"
	os.MkdirAll(path.Join(testDir, "shows"), 0755)
	store, _ := store.Open(testDir)
	defer store.Close()
	return searchHandler{store: store}
}

func cleanup() {
	os.RemoveAll("testdata/store")
}

func TestSearchWithoutQuery(t *testing.T) {
	defer cleanup()
	h := defaultSearchHandler()

	r, _ := http.NewRequest("GET", "http://example.com/search?q=", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected code 400, got:", w.Code)
	}
}

func TestSearchHandler_ServeHTTP(t *testing.T) {
	defer cleanup()
	h := defaultSearchHandler()

	r, _ := http.NewRequest("GET", "http://example.com/search?q=bla", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Body.String() != "hi bob!\n" {
		t.Errorf("unexpected response: %s", w.Body.String())
	}
}
