package torrents

var IsEnglish = isEnglish
var IsSeason = isSeason

var SearchEngines = searchEngines

func NewQueryJob(season int) queryJob {
	return queryJob{
		season: season,
	}
}
