package brave

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/karust/openserp/core"
)

func BuildURL(q core.Query) (string, error) {
	base, _ := url.Parse("https://search.brave.com/")
	base.Path += "search"

	params := url.Values{}
	if q.Text != "" {
		params.Add("q", q.Text)
	}

	if q.Page > 0 {
		params.Add("offset", strconv.Itoa(q.Page-1))
	}

	if len(params.Get("q")) == 0 {
		return "", errors.New("empty query built")
	}

	if q.DateInterval != "" {
		// Brave does not support date interval filtering via URL parameters.
		// return "", errors.New("date interval filtering not supported by Brave")
	}

	params.Add("spellcheck", "0")
	base.RawQuery = params.Encode()
	return base.String(), nil
}
