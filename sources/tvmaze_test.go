package sources_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/haarts/getme/sources"
	"github.com/haarts/getme/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTvMazeSearch(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/search/shows", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("testdata/tvmaze_search.json"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		require.False(t, true)
	})

	sources.SetTvMazeURL(ts.URL)

	results := (sources.TvMaze{}).Search("query")
	require.NoError(t, results.Error)
	require.Len(t, results.Shows, 10)
	assert.Equal(t, "Dead Set", results.Shows[0].Title)
}

func TestTvMazeSeasons(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/shows/1/episodes", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("testdata/tvmaze_episodes.json"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		require.False(t, true)
	})

	sources.SetTvMazeURL(ts.URL)

	seasons, err := (sources.TvMaze{}).Seasons(&store.Show{ID: 1})
	require.NoError(t, err)
	require.Len(t, seasons, 3)
	var season1 sources.Season
	for _, v := range seasons {
		if v.Season == 1 {
			season1 = v
		}
	}
	assert.Len(t, season1.Episodes, 13)
}

func readFixture(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
	return string(data)
}
