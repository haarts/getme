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
var fix bool

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s <flags>\n", os.Args[0])
		fmt.Println("Checks what is stored on disk and compares it to what is stored in storage.")
		fmt.Println("It will then tell what, in storage, is marked as NOT pending while no file on disk could be found.")
		fmt.Println("It has an option to change the storage to reflect the state on disk.\n")
		flag.PrintDefaults()
	}
	const (
		logLevelUsage = "Set log level (0,1,2,3,4,5, higher is more logging)."
		fixUsage      = "Fix the shows in storage."
	)

	flag.IntVar(&logLevel, "log-level", int(log.ErrorLevel), logLevelUsage)
	flag.IntVar(&logLevel, "l", int(log.ErrorLevel), logLevelUsage+" (shorthand)")

	flag.BoolVar(&fix, "fix", false, fixUsage)
	flag.BoolVar(&fix, "f", false, fixUsage+" (shorthand)")
}

func videosDir() (string, error) {
	if len(flag.Args()) != 1 {
		return "", errors.New("Expected an argument pointing to root dir containing videos")
	}

	dir := flag.Arg(0)
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
				markOrOutput(show, season.Season, episode)
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
				markOrOutput(show, season.Season, episode)
			}
		}
	}
}

func markOrOutput(show *store.Show, season int, episode *store.Episode) {
	if fix {
		episode.Pending = true
	} else {
		fmt.Printf(
			"'%s S%02dE%02d %s' missing on disk\n",
			show.Title,
			season,
			episode.Episode,
			episode.Title,
		)
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

	root, err := videosDir()
	if err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"args": flag.Args(),
		}).Fatal("Couldn't use argument as video location")
	}

	err = os.Chdir(root)
	if err != nil {
		log.WithFields(log.Fields{
			"err":  err,
			"root": root,
		}).Fatal("Error changing working dir")
	}

	store, err := store.Open(config.Config().StateDir)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error opening state")
	}
	defer store.Close()

	potentialShows, err := ioutil.ReadDir(root)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Error reading potential shows: ")
	}

	verifyShows(potentialShows, store)
}
