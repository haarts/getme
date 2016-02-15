package torrents_test

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/haarts/getme/torrents"
	"github.com/stretchr/testify/assert"
)

type mockDoner struct {
	doneFunc func()
}

func (m mockDoner) Done() {
	if m.doneFunc == nil {
		panic("Your test needs this, you should mock it.")
	}
	m.doneFunc()
}

func TestDownloadNonTorrent(t *testing.T) {
	mux, ts := Setup(t)
	defer ts.Close()

	mux.HandleFunc("/baz.torrent", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, "not really a torrent")
	})

	url, _ := url.Parse(ts.URL + "/baz.torrent")
	torrent := torrents.Torrent{
		URL:             url,
		Filename:        "baz",
		AssociatedMedia: mockDoner{},
	}

	torrents.Download([]torrents.Torrent{torrent}, "/tmp")

	// The torrent should NOT be stored
	_, err := os.Stat("/tmp/baz")
	fmt.Printf("err = %+v\n", err)
	assert.Error(t, err)
}
