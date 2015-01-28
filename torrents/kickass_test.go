package torrents

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/haarts/getme/store"
	"github.com/stretchr/testify/assert"
)

func readFixture(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
	return string(data)
}

func TestIsEnglish(t *testing.T) {
	assert := assert.New(t)

	ss := []SearchResult{
		{FileName: "it's all good"},
		{FileName: "this is very french"},
		{FileName: "some show vostfr"},
		{FileName: "some.show.ITA.avi"},
	}

	assert.True(isEnglish(ss[0]))

	for _, s := range ss[1:] {
		assert.False(isEnglish(s), "should not be english: %s", s.FileName)
	}
}

func TestSearching(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("testdata/kickass.xml"))
	}))
	defer ts.Close()

	kickassURL = ts.URL

	season := store.Season{1, []*store.Episode{{Pending: true, Episode: 1}}}
	show := store.Show{Title: "Title", URL: "url", Seasons: []*store.Season{&season}}
	matches, err := Search(&show)
	if err != nil {
		t.Error("Expected error to be nil, got:", err)
	}

	if len(matches) != 1 {
		t.Error("Expected matches to contain an equal about of entries as episodes searched for, got:", len(matches))
	}
}

func Test404(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	kickassURL = ts.URL

	season := store.Season{1, []*store.Episode{{Pending: true, Episode: 1}}}
	show := store.Show{Title: "Title", URL: "url", Seasons: []*store.Season{&season}}
	matches, err := Search(&show)

	if err != nil {
		t.Error("Not finding a torrent is not a big deal. Just continue. Got:", err)
	}
	if len(matches) != 0 {
		t.Error("Not finding a torrent is not a big deal. Just continue. Got:", matches)
	}
}
