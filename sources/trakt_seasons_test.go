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

	seasons, _ := GetSeasons(Match{URL: "boo/some-url"})
	if len(seasons) != 6 {
		t.Error("Expected 6 seasons, got:", len(seasons))
	}

	if seasons[0].Season == 0 {
		t.Error("Expected Season field to be not default, got:", seasons[0])
	}
}
