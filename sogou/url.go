package sogou

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/karust/openserp/core"
)

func BuildURL(q core.Query) (string, error) {
	base, _ := url.Parse("https://www.sogou.com/")
	base.Path += "web"

	text := strings.TrimSpace(q.Text)
	if q.Site != "" {
		text += " site:" + q.Site
	}
	if q.Filetype != "" {
		text += " filetype:" + q.Filetype
	}

	params := url.Values{}
	if text != "" {
		params.Add("query", text)
	}

	if q.Page > 0 {
		params.Add("page", strconv.Itoa(q.Page))
	}

	if len(params.Get("query")) == 0 {
		return "", errors.New("empty query built")
	}

	params.Add("ie", "utf8")
	base.RawQuery = params.Encode()
	return base.String(), nil
}
