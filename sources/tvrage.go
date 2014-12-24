package sources

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func init() {
	Register("tvrage", searchTvRage)
}

type tvRageResult struct {
	Shows []tvRageMatch `xml:"show"`
}
type tvRageMatch struct {
	ID    int    `xml:"id"`
	Title string `xml:"name"`
}

var tvRageSearchURL = "http://services.tvrage.com"

func constructTvRageURL(query string) string {
	return fmt.Sprintf(tvRageSearchURL+"/feeds/search.php?show=%s", url.QueryEscape(query))
}

func searchTvRage(query string) ([]Match, error) {
	resp, err := http.Get(constructTvRageURL(query))
	if err != nil {
		return nil, err //TODO retry a couple of times when it's a timeout.
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Search returned non 200 status code: %d", resp.StatusCode))
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result tvRageResult
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return convertTvRageToMatches(result.Shows), nil
}

func convertTvRageToMatches(ms []tvRageMatch) []Match {
	matches := make([]Match, len(ms))
	for i, m := range ms {
		matches[i] = Show{
			Title: m.Title,
			ID:    m.ID,
			seasonsAndEpisodesFunc: getSeasonsOnTvRage,
		}
	}
	return matches
}

func getSeasonsOnTvRage(show *Show) error {
	return nil
}
