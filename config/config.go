//Package config handles GetMe's configuration.
package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// Config contains WHERE downloaded torrents should be copied too. And WHERE
// the state should be stored. And WHERE the log files should be stored.
type Conf struct {
	WatchDir, StateDir, LogDir string
}

// CheckConfig see if the config file is present.
func CheckConfig() error {
	_, err := os.Stat(ConfigFile())
	return err
}

var memoizedConfig *Conf
var failed bool

func SetLoggerOutput(logDir string) {
	f, err := os.OpenFile(path.Join(logDir, "getme.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Something went wrong opening the logfile:", err)
	}
	log.SetOutput(f)
}

func SetLoggerTo(logLevel int) {
	log.SetLevel(log.Level(logLevel))
}

// Config returns a config object.
func Config() *Conf {
	if memoizedConfig != nil {
		return memoizedConfig
	}
	if failed {
		return nil
	}

	// read config file
	file, err := os.Open(ConfigFile())
	if err != nil {
		fmt.Println("Something went wrong reading the config file:", err)
		failed = true
		return nil
	}
	defer file.Close()

	conf := Conf{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		parts := strings.Split(text, "=")
		for i := range parts {
			parts[i] = strings.Trim(parts[i], " ")
		}
		switch parts[0] {
		case "watch_dir":
			conf.WatchDir = parts[1]
		default:
			fmt.Println("Found an unknown key in config.ini: " + parts[0])
			failed = true
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Something went wrong reading the config file:", err) //TODO replace with log.Fatal()
		failed = true
		return nil
	}

	// setup watch dir
	err = ensureWatchDir(conf.WatchDir)
	if err != nil {
		fmt.Println("Something went wrong creating the watch directory:", err)
		failed = true
		return nil
	}

	// setup logging dir
	err = ensureLogDir(logDir())
	if err != nil {
		fmt.Println("Something went wrong creating the log directories:", err)
		failed = true
		return nil
	}
	conf.LogDir = logDir()

	// setup storage/state dirs
	conf.StateDir = stateDir()
	err = ensureStateDir(conf.StateDir)
	if err != nil {
		fmt.Println("Something went wrong creating the state directories:", err) //TODO replace with log.Fatal()
		failed = true
		return nil
	}

	memoizedConfig = &conf
	return memoizedConfig
}

func ensureWatchDir(watchDir string) error {
	return ensureDirs([]string{watchDir})
}

func ensureLogDir(logDir string) error {
	return ensureDirs([]string{logDir})
}

func ensureStateDir(stateDir string) error {
	dirs := []string{
		stateDir,
		path.Join(stateDir, "shows"),
		path.Join(stateDir, "movies"),
	}

	return ensureDirs(dirs)
}

func ensureDirs(dirs []string) error {
	for _, d := range dirs {
		err := os.MkdirAll(d, 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func userHomeDir() string {
	var u *user.User
	if u, _ = user.Current(); u == nil {
		return "" // TODO handle err
	}
	return u.HomeDir
}

// NOTE this is not XGD standard but suggested by Debian. See:
// https://stackoverflow.com/questions/25897836/where-should-i-write-a-user-specific-log-file-to-and-be-xdg-base-directory-comp/27965014#27965014
func logDir() string {
	dir := path.Join(userHomeDir(), ".local", "state", "getme") // TODO What's the sane location for Windows?
	return dir
}

func stateDir() string {
	dir := path.Join(userHomeDir(), ".local", "share", "getme") // TODO What's the sane location for Windows?
	return dir
}

func ConfigFile() string {
	dirPath := path.Join(userHomeDir(), ".config", "getme") // TODO What's the sane location for Windows?
	filePath := path.Join(dirPath, "config.ini")
	return filePath
}

func WriteDefaultConfig() {
	f := ConfigFile()
	dir, _ := path.Split(f)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}
	if err := ioutil.WriteFile(f, defaultConfigData(dir), 0644); err != nil {
		return
	}
}

func defaultConfigData(homeDir string) []byte {
	watchDir := fmt.Sprintln("watch_dir = /tmp/torrents")
	return []byte(watchDir)
}
