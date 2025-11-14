package brave

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/corpix/uarand"
	"github.com/karust/openserp/core"
	"github.com/sirupsen/logrus"

	utls "github.com/refraction-networking/utls"
)

func braveRequest(searchURL string, query core.Query) (*http.Response, error) {
	// Create HTTP transport with proxy
	transport := &http.Transport{}
	if query.ProxyURL != "" {
		proxyUrl, err := url.Parse(query.ProxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	// Set insecure TLS
	if query.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	transport.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{}
		rawConn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		hostname := strings.Split(addr, ":")[0]
		config := &utls.Config{
			ServerName:         hostname,
			InsecureSkipVerify: query.Insecure,
		}

		uconn := utls.UClient(rawConn, config, utls.HelloChrome_Auto)

		if err := uconn.Handshake(); err != nil {
			rawConn.Close()
			return nil, err
		}

		return uconn, nil
	}

	baseClient := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 10,
	}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", uarand.GetRandom())

	res, err := baseClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func braveResultParser(response *http.Response) ([]core.SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	results := []core.SearchResult{}
	rank := 1

	doc.Find("div.snippet").Each(func(i int, s *goquery.Selection) {
		titleTag := s.Find("a.title")
		link, _ := titleTag.Attr("href")
		title := titleTag.Text()
		description := s.Find("div.description").Text()

		if link != "" && link != "#" {
			result := core.SearchResult{
				Rank:        rank,
				URL:         link,
				Title:       title,
				Description: description,
			}
			results = append(results, result)
			rank++
		}
	})

	logrus.Tracef("Brave search document size: %d", len(doc.Text()))
	return results, err
}

func SearchRaw(query core.Query) ([]core.SearchResult, error) {
	braveURL, err := BuildURL(query)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Brave URL built: %s", braveURL)

	res, err := braveRequest(braveURL, query)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Brave Raw response: code=%d", res.StatusCode)

	results, err := braveResultParser(res)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Brave Raw results : %v", results)

	return core.DeduplicateResults(results), nil
}
