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

	show := &Show{ID: 123, SourceName: tvRageName}
	GetSeasonsAndEpisodes(show)

	if len(show.Seasons) == 0 {
		t.Fatal("Expected seasons to be not zero, got:", len(show.Seasons))
	}

	season := show.Seasons[0]
	if season.Season == 0 {
		t.Error("Expected Season number not to be zero, got:", season.Season)
	}
}

func TestTvRageSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, readFixture("fixtures/tvrage_search.xml"))
	}))
	defer ts.Close()

	tvRageURL = ts.URL

	matches, _ := (TvRage{}).Search("some query")
	if matches[0].DisplayTitle() != "The Big Bang Theory" {
		t.Error("Best match is not The Big Bang Theory")
	}

	s := (matches[0]).(Show)
	if s.ID == 0 {
		t.Error("Expect ID to be not zero, got:", s.ID)
	}
}
