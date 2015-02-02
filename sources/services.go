package sources

func popularShowAtIndex(shows []Show) int {
	for i, s := range shows {
		if isPopularShow(s.Title) {
			return i
		}
	}

	return -1
}

func isPopularShow(showName string) bool {
	for _, n := range SHOWS {
		if showName == n {
			return true
		}
	}
	return false
}
