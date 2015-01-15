//config handles GetMe's configuration.
package config

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
)

// Config contains WHERE downloaded torrents should be copied too. And WHERE
// the state should be stored. And WHERE the log files should be stored.
type Conf struct {
	WatchDir string
	StateDir string
	Logger   *logrus.Logger
}

func CheckConfig() error {
	_, err := os.Stat(ConfigFile())
	return err
}

var memoizedConfig *Conf

func Config() *Conf {
	if memoizedConfig != nil {
		return memoizedConfig
	}
	file, err := os.Open(ConfigFile())
	if err != nil {
		fmt.Println("Something went wrong reading the config file:", err) //TODO replace with log.Fatal()
		return nil
	}
	defer file.Close()

	f, err := os.Open(path.Join(logDir(), "getme.log"))
	if err != nil {
		fmt.Println("Something went wrong opening the logfile file:", err) //TODO replace with log.Fatal()
		return nil
	}

	log := logrus.New()
	log.Out = f
	conf := Conf{
		Logger:   log,
		StateDir: stateDir(),
	}
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
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Something went wrong reading the config file:", err) //TODO replace with log.Fatal()
		os.Exit(1)
	}
	memoizedConfig = &conf
	return memoizedConfig
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
