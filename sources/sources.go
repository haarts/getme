package sources

type Matches interface {
	BestMatch() Match
}

type Match interface {
	Title() string
}
