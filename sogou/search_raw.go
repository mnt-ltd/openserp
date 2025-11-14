package sogou

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

func sogouRequest(searchURL string, query core.Query) (*http.Response, error) {
	transport := &http.Transport{}
	if query.ProxyURL != "" {
		proxyUrl, err := url.Parse(query.ProxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

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

	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", uarand.GetRandom())

	return client.Do(req)
}

func sogouResultParser(response *http.Response) ([]core.SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	results := []core.SearchResult{}
	rank := 1

	doc.Find("div.vrwrap, div.rb").Each(func(i int, s *goquery.Selection) {
		titleTag := s.Find("h3 a")
		if titleTag.Length() == 0 {
			titleTag = s.Find("a.pt")
		}
		link, _ := titleTag.Attr("href")
		title := strings.TrimSpace(titleTag.Text())
		description := strings.TrimSpace(s.Find("div.ft").Text())

		if link != "" && link != "#" {
			if !(strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://")) {
				link = "http://www.sogou.com" + link
			}
			results = append(results, core.SearchResult{
				Rank:        rank,
				URL:         link,
				Title:       title,
				Description: description,
			})
			rank++
		}
	})

	logrus.Tracef("Sogou search document size: %d", len(doc.Text()))
	return results, nil
}

func SearchRaw(query core.Query) ([]core.SearchResult, error) {
	sogouURL, err := BuildURL(query)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Sogou URL built: %s", sogouURL)

	res, err := sogouRequest(sogouURL, query)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Sogou Raw response: code=%d", res.StatusCode)

	results, err := sogouResultParser(res)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Sogou Raw results : %v", results)

	return core.DeduplicateResults(results), nil
}
