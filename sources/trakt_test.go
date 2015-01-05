package sources

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func readFixture(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
	return string(data)
}

func TestOrdering(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("fixtures/trakt_search.json"))
	}))
	defer ts.Close()
	defer func() {
		sources = make(map[string]Source)
		Register(TVRAGE, TvRage{})
		Register(TRAKT, Trakt{})
	}()

	traktSearchURL = ts.URL + "/search?type=show"

	sources = make(map[string]Source)
	Register(TRAKT, Trakt{})
	matches, _ := Search("some query")
	if matches[0][0].DisplayTitle() != "Game of Thrones" {
		t.Fatal("best match is not Game of Thrones, got:", matches[0][0].DisplayTitle())
	}
}
