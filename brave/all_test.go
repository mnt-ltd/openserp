package brave_test

import (
	"testing"

	"github.com/karust/openserp/brave"
	"github.com/karust/openserp/core"
	"github.com/sirupsen/logrus"
)

var browser *core.Browser
var searchEngine *brave.BraveSearch

func init() {
	var err error
	opts := core.BrowserOpts{IsHeadless: true}
	browser, err = core.NewBrowser(opts)
	if err != nil {
		logrus.Fatal(err)
	}

	searchEngine = brave.New(*browser, core.SearchEngineOptions{})
}

func TestSearch(t *testing.T) {
	query := core.Query{Text: "cat"}
	results, err := searchEngine.Search(query)
	if err != nil {
		t.Error(err)
	}

	if len(results) == 0 {
		t.Error("No results found")
	}

	for _, res := range results {
		if res.Title == "" || res.URL == "" {
			t.Errorf("Empty title or URL in result: %+v", res)
		}
		t.Logf("%+v", res)
	}
}
