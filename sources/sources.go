package sources

type Match struct {
	Title string
	URL   string // NOTE: Perhaps this should be a url.URL in stead of a string
}

type Season struct {
	Season   int
	Episodes int
}
