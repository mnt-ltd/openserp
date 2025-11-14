package brave

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/karust/openserp/core"
	"golang.org/x/time/rate"
)

type BraveSearch struct {
	core.Browser
	core.SearchEngineOptions
	logger *core.EngineLogger
}

func New(browser core.Browser, opts core.SearchEngineOptions) *BraveSearch {
	s := BraveSearch{Browser: browser}
	opts.Init()
	s.SearchEngineOptions = opts
	s.logger = core.NewEngineLogger("Brave")
	return &s
}

func (s *BraveSearch) Name() string {
	return "brave"
}

func (s *BraveSearch) GetRateLimiter() *rate.Limiter {
	ratelimit := rate.Every(s.GetRatelimit())
	return rate.NewLimiter(ratelimit, s.RateBurst)
}

func (s *BraveSearch) isCaptcha(page *rod.Page) bool {
	_, err := page.Timeout(s.GetSelectorTimeout()).Search("div.h-captcha")
	return err == nil
}

func (s *BraveSearch) Search(query core.Query) ([]core.SearchResult, error) {
	s.logger.Debug("Starting search, query: %+v", query)

	searchResults := []core.SearchResult{}

	url, err := BuildURL(query)
	if err != nil {
		return nil, err
	}

	page, err := s.Navigate(url)
	if err != nil {
		return nil, err
	}

	results, err := page.Timeout(s.Timeout).Search("div.snippet")
	if err != nil {
		defer page.Close()
		s.logger.Error("Cannot parse search results: %s", err)
		return nil, core.ErrSearchTimeout
	}

	if results == nil {
		defer page.Close()

		if s.isCaptcha(page) {
			s.logger.Error("Captcha detected: %s", url)
			return nil, core.ErrCaptcha
		}
		s.logger.Info("No results found on page: %s", url)
		return searchResults, nil
	}

	elements, err := results.All()
	if err != nil {
		defer page.Close()
		s.logger.Error("Cannot get all search results: %s", err)
		return nil, err
	}

	s.logger.Debug("Parsing %d search results", len(elements))

	for i, r := range elements {
		title, err := r.Element("a.heading-serpresult")
		if err != nil {
			s.logger.Warn("Cannot get title for #%d result: %s", i, err)
			continue
		}

		link, err := title.Attribute("href")
		if err != nil {
			s.logger.Warn("Cannot get link for #%d result: %s", i, err)
			continue
		}

		snippet, err := r.Element("div.snippet-description")
		if err != nil {
			s.logger.Warn("Cannot get snippet for #%d result: %s", i, err)
			continue
		}

		titleText, _ := title.Text()
		snippetText, _ := snippet.Text()

		searchResult := core.SearchResult{
			Rank:        len(searchResults) + 1,
			URL:         *link,
			Title:       titleText,
			Description: snippetText,
		}
		searchResults = append(searchResults, searchResult)
	}

	defer page.Close()
	return searchResults, nil
}

func (s *BraveSearch) SearchImage(query core.Query) ([]core.SearchResult, error) {
	return nil, fmt.Errorf("image search is not supported for brave")
}
