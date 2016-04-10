package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/haarts/getme/config"
	"github.com/haarts/getme/store"
)

var logLevel int

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s <flags>\n", os.Args[0])
		flag.PrintDefaults()
	}
	const (
		logLevelUsage = "Set log level (0,1,2,3,4,5, higher is more logging)."
	)

	flag.IntVar(&logLevel, "log-level", int(log.ErrorLevel), logLevelUsage)
	flag.IntVar(&logLevel, "l", int(log.ErrorLevel), logLevelUsage+" (shorthand)")
}

func videosDir() (string, error) {
	if len(flag.Args()) != 1 {
		return "", errors.New("Expected an argument pointing to root dir containing videos")
	}

	dir := flag.Arg(0)
	fmt.Printf("dir = %+v\n", dir)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", err
	}

	return dir, nil
}

func matchDirWithShow(potentialShow os.FileInfo, store *store.Store) *store.Show {
	contextLogger := log.WithField("file", potentialShow.Name())
	if !potentialShow.IsDir() {
		contextLogger.Info("File is not a directory and therefor not a show")
		return nil
	}

	for _, show := range store.Shows() {
		if show.Title == potentialShow.Name() {
			contextLogger.WithField("show", show.Title).Info("Matched a show")
			return show
		}
	}

	contextLogger.Warn("Failed to match directory to show")
	return nil
}

func verifyPendingStates(dir os.FileInfo, show *store.Show) {
	contextLogger := log.WithField("dir", dir.Name())

	for _, season := range show.Seasons {
		contextLogger = contextLogger.WithField("season", season.Season)

		pathPrefix := path.Join(dir.Name(), strconv.Itoa(season.Season))

		for _, episode := range season.Episodes {
			contextLogger = contextLogger.WithField("episode", episode.Episode)

			if episode.Pending {
				contextLogger.Debug("Skipping because episode is already pending")
				continue
			}
			expectedVideoName := fmt.Sprintf("S%02dE%02d", season.Season, episode.Episode)

			files, err := ioutil.ReadDir(pathPrefix)
			if err != nil {
				contextLogger.Info("Marking episode as pending as ReadDir failed")
				episode.Pending = true
				continue
			}

			found := false
			for _, file := range files {
				if strings.Contains(file.Name(), expectedVideoName) {
					contextLogger.Debug("Found a file on disk matching the episode")
					found = true
					break
				}
			}
			if !found {
				contextLogger.Info("Marking episode as pending as no match was found on disk")
				episode.Pending = true
			}
		}
	}
}

func verifyShows(potentialShows []os.FileInfo, store *store.Store) {
	for _, potentialShow := range potentialShows {
		show := matchDirWithShow(potentialShow, store)
		if show == nil {
			log.WithField("dir", potentialShow.Name()).Info("Skipping dir")
			continue
		}
		verifyPendingStates(potentialShow, show)
	}
}

func main() {
	flag.Parse()
	log.SetLevel(log.Level(logLevel))

	dir, err := videosDir()
	if err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"args": flag.Args(),
		}).Fatal("Couldn't use argument as video location")
	}

	store, err := store.Open(config.Config().StateDir)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error opening state")
	}
	defer store.Close()

	potentialShows, err := ioutil.ReadDir(dir)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error reading potential shows: ")
	}

	verifyShows(potentialShows, store)
}
