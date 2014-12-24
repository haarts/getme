package sources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSeasons(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("fixtures/trakt_seasons.json"))
	}))
	defer ts.Close()

	traktSeasonsURL = ts.URL + "/"

	show := &Show{URL: "boo/some-url"}
	getSeasonsOnTrakt(show)
	if len(show.Seasons) != 6 {
		t.Error("Expected 6 seasons, got:", len(show.Seasons))
	}

	if show.Seasons[0].Season == 0 {
		t.Error("Expected Season field to be not default, got:", show.Seasons[0])
	}
}
