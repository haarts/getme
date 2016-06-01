package torrents_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/haarts/getme/torrents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTorrentProjectSearch(t *testing.T) {
	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "s=foo&out=json&orderby=seeds", r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, ReadFixture("testdata/torrentproject.json"))
	})

	torrentProject := torrents.TorrentProject{
		URL:         ts.URL,
		TorCacheURL: "%s",
	}

	results, err := torrentProject.Search("foo")
	require.NoError(t, err)
	require.Len(t, results, 10)
	assert.Equal(t, "Udemy - Ubuntu Desktop for Beginners - Start Using Linux Today!", results[0].Title)
}
