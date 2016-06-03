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

func TestExtraTorrentSearch(t *testing.T) {
	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	http.DefaultServeMux.HandleFunc("/search/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "search=foo&new=1&x=0&y=0", r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, ReadFixture("testdata/extratorrent.xml"))
	})

	engine := torrents.ExtraTorrent{
		URL: ts.URL,
	}

	results, err := engine.Search("foo")
	require.NoError(t, err)
	require.Len(t, results, 3)
	assert.Equal(t, "One Flew Over The Cuckoos Nest (1975) 720p MKV x264 AC3 BRrip [Pioneer]", results[0].Title)
}
