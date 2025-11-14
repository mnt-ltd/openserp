package sogou

import (
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/karust/openserp/core"
	"golang.org/x/time/rate"
)

type Sogou struct {
	core.Browser
	core.SearchEngineOptions
	logger *core.EngineLogger
}

func New(browser core.Browser, opts core.SearchEngineOptions) *Sogou {
	s := Sogou{Browser: browser}
	opts.Init()
	s.SearchEngineOptions = opts
	s.logger = core.NewEngineLogger("Sogou")
	return &s
}

func (s *Sogou) Name() string {
	return "sogou"
}

func (s *Sogou) GetRateLimiter() *rate.Limiter {
	ratelimit := rate.Every(s.GetRatelimit())
	return rate.NewLimiter(ratelimit, s.RateBurst)
}

func (s *Sogou) isCaptcha(page *rod.Page) bool {
	_, err := page.Timeout(s.GetSelectorTimeout()).Search(".vr-captcha, #seccode")
	return err == nil
}

func (s *Sogou) Search(query core.Query) ([]core.SearchResult, error) {
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

	results, err := page.Timeout(s.Timeout).Search("div.vrwrap, div.rb")
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
		title, err := r.Element("h3 a")
		if err != nil {
			// try alternative selector
			title, err = r.Element("a.pt")
			if err != nil {
				s.logger.Warn("Cannot get title for #%d result: %s", i, err)
				continue
			}
		}

		link, err := title.Attribute("href")
		if err != nil || link == nil {
			s.logger.Warn("Cannot get link for #%d result: %v", i, err)
			continue
		}

		snippet, _ := r.Element("div.ft")
		titleText, _ := title.Text()
		snippetText := ""
		if snippet != nil {
			snippetText, _ = snippet.Text()
		}

		if !(strings.HasPrefix(*link, "http://") || strings.HasPrefix(*link, "https://")) {
			*link = "http://www.sogou.com" + *link
		}

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

func (s *Sogou) SearchImage(query core.Query) ([]core.SearchResult, error) {
	return nil, fmt.Errorf("image search is not supported for sogou")
}
