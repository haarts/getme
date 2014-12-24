package sources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTvRageSeasons(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, readFixture("fixtures/tvrage_seasons.xml"))
	}))
	defer ts.Close()

	tvRageURL = ts.URL

	show := &Show{ID: 123}
	getSeasonsOnTvRage(show)
	fmt.Printf("show %+v\n", show)
}

func TestTvRageSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, readFixture("fixtures/tvrage_search.xml"))
	}))
	defer ts.Close()

	tvRageURL = ts.URL

	matches, _ := searchTvRage("some query")
	if matches[0].DisplayTitle() != "The Big Bang Theory" {
		t.Fatal("best match is not The Big Bang Theory")
	}
}
