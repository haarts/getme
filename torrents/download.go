package torrents

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
)

var timeout = 2 * time.Second
var requestDelay = 5 * time.Second

// Download takes a slice of torrents and downloads them to destination.
// It rate limits the requests per host. And times requests out after a
// while.
func Download(foundTorrents []Torrent, destination string) error {
	tickers := map[string]<-chan time.Time{}
	relevantTicker := func(host string) <-chan time.Time {
		if ticker, ok := tickers[host]; ok {
			return ticker
		}
		tickers[host] = time.Tick(requestDelay)
		return tickers[host]
	}

	errors := make(chan error)
	for _, foundTorrent := range foundTorrents {
		go func(t Torrent) {
			<-relevantTicker(t.URL.Host) // rate limit ourselves
			err := downloadWithTimeout(t, destination)
			if err == nil {
				log.WithFields(log.Fields{
					"torrent": t.URL,
				}).Debug("Download successful")

				t.AssociatedMedia.Done()
			}
			errors <- err
		}(foundTorrent)
	}

	var err error
	for i := 0; i < len(foundTorrents); i++ {
		err = <-errors
	}
	return err
}

func downloadWithTimeout(torrent Torrent, destination string) error {
	result := make(chan error)
	go func() {
		result <- download(torrent, destination)
	}()

	select {
	case <-time.After(timeout):
		return fmt.Errorf("download timed out on '%s'", torrent.URL)
	case err := <-result:
		return err
	}
}

func download(torrent Torrent, destination string) error {
	output, err := os.Create(path.Join(destination, torrent.Filename))
	if err != nil {
		log.WithFields(log.Fields{
			"torrent": torrent.Filename,
			"err":     err,
		}).Warn("File creation failed")
		return err
	}
	defer output.Close()

	cleanup := func() error {
		stat, err := output.Stat()
		if err != nil {
			return err
		}
		return os.Remove(path.Join(destination, stat.Name()))
	}

	response, err := http.Get(torrent.URL.String())
	if err != nil {
		log.WithFields(log.Fields{
			"torrent": torrent.Filename,
			"err":     err,
		}).Warn("Download failed")
		_ = cleanup()
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"torrent": torrent.Filename,
			"err":     err,
		}).Warn("Copy failed")
		_ = cleanup()
		return err
	}

	return nil
}
