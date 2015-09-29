package torrents_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/haarts/getme/torrents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTorrentCDSearch(t *testing.T) {
	mux, ts := setup(t)
	defer ts.Close()

	mux.HandleFunc("/torrents/xml", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "q=foo", r.URL.RawQuery)

		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, ReadFixture("testdata/torrentcd.xml"))
	})

	torrentCD := torrents.TorrentCD{
		URL: ts.URL,
	}

	results, err := torrentCD.Search("foo")
	require.NoError(t, err)
	require.Len(t, results, 10)
	assert.Equal(t, "foo bar", results[0].OriginalName)
}

func setup(t *testing.T) (*http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		require.False(t, true, "Catch all invoked with %s", r.URL.String())
	})

	return mux, ts
}

func ReadFixture(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
	return string(data)
}
