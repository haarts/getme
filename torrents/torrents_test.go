package torrents_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/haarts/getme/store"
	"github.com/haarts/getme/torrents"
)

func TestSearchForTorrents(t *testing.T) {
	mux, ts := Setup(t)
	defer ts.Close()

	mux.HandleFunc("/usearch/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, ReadFixture("testdata/kickass.xml"))
	})

	torrents.SearchEngines["kickass"] = torrents.Kickass{
		URL: ts.URL,
	}
	delete(torrents.SearchEngines, "torrentcd")
	delete(torrents.SearchEngines, "extratorrent")

	season := store.Season{1, []*store.Episode{{Pending: true, Episode: 1}}}
	show := store.Show{Title: "Title", URL: "url", Seasons: []*store.Season{&season}}
	matches, err := torrents.Search(&show)
	require.NoError(t, err)

	assert.Equal(t, 1, len(matches))
}

func Test404(t *testing.T) {
	mux, ts := Setup(t)
	defer ts.Close()

	mux.HandleFunc("/usearch/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	torrents.SearchEngines["kickass"] = torrents.Kickass{URL: ts.URL}
	delete(torrents.SearchEngines, "torrentcd")
	delete(torrents.SearchEngines, "extratorrent")

	season := store.Season{1, []*store.Episode{{Pending: true, Episode: 1}}}
	show := store.Show{Title: "Title", URL: "url", Seasons: []*store.Season{&season}}
	matches, err := torrents.Search(&show)
	require.NoError(t, err, "Not finding a torrent is not a big deal. Just continue.")

	assert.Equal(t, 0, len(matches))
}

func TestIsEnglish(t *testing.T) {
	ss := []string{
		"it's all good",
		"this is very french",
		"some show vostfr",
		"some.show.ITA.avi",
	}

	assert.True(t, torrents.IsEnglish(ss[0]))

	for _, s := range ss[1:] {
		assert.False(t, torrents.IsEnglish(s), "should not be english: %s", s)
	}
}
