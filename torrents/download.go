package torrents

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jackpal/bencode-go"
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

func download(torrent Torrent, directory string) error {
	logEntry := log.WithFields(log.Fields{
		"torrent": torrent.Filename,
	})

	response, err := http.Get(torrent.URL.String())
	if err != nil {
		logEntry.WithFields(log.Fields{
			"err": err,
		}).Warn("Download failed")
		return err
	}
	defer response.Body.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		logEntry.WithFields(log.Fields{
			"err": err,
		}).Warn("Reading response body failed")
		return err
	}

	_, err = bencode.Decode(buf)
	if err != nil {
		logEntry.WithFields(log.Fields{
			"err": err,
		}).Warn("Torrent could not be decoded")
		return err
	}

	file, err := os.Create(path.Join(directory, torrent.Filename))
	if err != nil {
		logEntry.WithFields(log.Fields{
			"err": err,
		}).Warn("File creation failed")
		return err
	}
	defer file.Close()

	cleanup := func() error {
		stat, err := file.Stat()
		if err != nil {
			return err
		}
		return os.Remove(path.Join(directory, stat.Name()))
	}

	_, err = io.Copy(file, buf)
	if err != nil {
		logEntry.WithFields(log.Fields{
			"err": err,
		}).Warn("Copy to file failed")
		_ = cleanup()
		return err
	}

	return nil
}
