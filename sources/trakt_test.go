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

	traktSearchURL = ts.URL

	matches, _ := Search("some query")
	if matches[0][0].DisplayTitle() != "The Big Bang Theory" {
		t.Fatal("best match is not The Big Bang Theory")
	}
}
