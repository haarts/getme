package sources

import (
	"time"

	"github.com/42minutes/go-trakt"
	"github.com/haarts/getme/store"
)

// Trakt is the struct which implements the Source interface.
type Trakt struct{}

const traktName = "trakt"
const traktURL = "https://api-v2launch.trakt.tv"

func (t Trakt) Name() string {
	return traktName
}

func (t Trakt) Seasons(show *store.Show) ([]Season, error) {
	var seasons []Season

	client := traktClient()
	traktSeasons, result := client.Seasons().All(show.ID)
	if result.Err != nil {
		return seasons, result.Err
	}

	for i := 0; i < len(traktSeasons); i++ {
		season := Season{
			Season: traktSeasons[i].Number,
		}
		episodes, result := client.Episodes().AllBySeason(show.ID, traktSeasons[i].Number)
		if result.Err != nil {
			return seasons, result.Err
		}
		for _, episode := range episodes {
			if episode.FirstAired == nil {
				episode.FirstAired = &time.Time{}
			}
			season.Episodes = append(
				season.Episodes,
				Episode{
					Title:   episode.Title,
					AirDate: *episode.FirstAired,
					Episode: episode.Number,
				},
			)
		}
		seasons = append(seasons, season)
	}
	return seasons, nil
}

func (t Trakt) Search(q string) SearchResult {
	searchResult := SearchResult{
		Name: traktName,
	}

	results, response := traktClient().Shows().Search(q)
	if response.Err != nil {
		searchResult.Error = response.Err
		return searchResult
	}

	for _, result := range results {
		ended := t.isEnded(result.Show.Status)
		searchResult.Shows = append(
			searchResult.Shows,
			Show{
				Title:  result.Show.Title,
				ID:     result.Show.IDs.Trakt,
				Ended:  &ended,
				URL:    traktURL + "shows/" + result.Show.IDs.Slug,
				Source: searchResult.Name,
			},
		)
	}
	return searchResult
}

// part of the value t because of namespace conflict.
func (t Trakt) isEnded(status string) bool {
	return status == "ended"
}

func traktClient() *trakt.Client {
	apiKey := "01045164ed603042b53acf841b590f0e7b728dbff319c8d128f8649e2427cbe9"
	authMethod := trakt.TokenAuth{AccessToken: "3b6f5bdba2fa56b086712d5f3f15b4e967f99ab049a6d3a4c2e56dc9c3c90462"}

	return trakt.NewClientWith(
		traktURL,
		trakt.UserAgent,
		apiKey,
		authMethod,
		nil,
	)
}
