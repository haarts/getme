package search_engines

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/haarts/getme/sources"
)

func readFixture(file string) string {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
	return string(data)
}

func TestSearching(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, readFixture("fixtures/kickass.xml"))
	}))
	defer ts.Close()

	kickassSearchURL = ts.URL

	matches, err := Search([]sources.Episode{{ShowName: "Bar", Episode: 1, Season: 1}})
	fmt.Printf("err %+v\n", err)
	fmt.Printf("matches %+v\n", matches)
}
