package sources_test

import (
	"testing"

	"github.com/haarts/getme/sources"
)

func searchTestFunction(_ string) ([]sources.Match, error) {
	return make([]sources.Match, 0), nil
}

func TestRegisterDuplicateSource(t *testing.T) {
	defer func() {
		str := recover()
		if str == nil {
			t.Error("Expected panic, got none.")
		}
	}()
	sources.Register("one", searchTestFunction)
	sources.Register("one", searchTestFunction)
}
