package sources_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/haarts/getme/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTvMazeSearch(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/search/shows", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("MATCH")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("testdata/tvmaze_search.json"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("NOOS")
	})

	sources.SetTvMazeURL(ts.URL)

	results := (sources.TvMaze{}).Search("query")
	require.NoError(t, results.Error)
	require.Len(t, results.Shows, 10)
	assert.Equal(t, "Dead Set", results.Shows[0].Title)
}

func TestTvMazeSeasons(t *testing.T) {
	//ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/json")
	//fmt.Fprintln(w, readFixture("testdata/tvmaze_search.json"))
	//}))
	//defer ts.Close()

	//tvMazeURL = ts.URL

	//seasons, _ := (TvMaze{}).Seasons("some query")

}

func readFixture(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
	return string(data)
}
