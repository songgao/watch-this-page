package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/publicsuffix"
)

type watcher struct {
	u     *url.URL
	query string

	client       *http.Client
	lastModified string
	etag         string
}

func newWatcher(pageURL string, query string) (*watcher, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}

	return &watcher{
		u:     u,
		query: query,
		client: &http.Client{
			Jar: jar,
		},
	}, nil
}

type cookies []*http.Cookie

func (c cookies) MarshalZerologObject(e *zerolog.Event) {
	for _, cookie := range c {
		e.Str("cookie:"+cookie.Name, cookie.Value)
	}
}

func (f *watcher) makeRequest() (*http.Request, error) {
	req, err := http.NewRequest("GET", f.u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36 Edg/103.0.1264.51")

	if len(f.lastModified) > 0 {
		req.Header.Add("if-modified-since", f.lastModified)
	}
	if len(f.etag) > 0 {
		req.Header.Add("if-none-match", f.etag)
	}

	return req, nil
}

func (f *watcher) getWatchedField() (target string, changed bool, err error) {
	req, err := f.makeRequest()
	if err != nil {
		log.Error().Str("stage", "fetch").Err(err).Msg("making http request error")
		return "", false, err
	}
	res, err := f.client.Do(req)
	if err != nil {
		log.Error().Str("stage", "fetch").Err(err).Msg("fetch error")
		return "", false, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNotModified {
		log.Error().Str("stage", "fetch").
			Int("http_status_code", res.StatusCode).
			Str("http_status", res.Status).
			Msg("fetch error")
		return "", false, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	changed = res.Header.Get("last-modified") != f.lastModified || res.Header.Get("etag") != f.etag
	f.lastModified = res.Header.Get("last-modified")
	f.etag = res.Header.Get("etag")

	log.Debug().Str("stage", "fetch").
		Object("cookies", cookies(f.client.Jar.Cookies(f.u))).
		Str("last-modified", f.lastModified).
		Str("etag", f.etag).
		Int("status_code", res.StatusCode).
		Msg("GET success")

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Error().Str("stage", "fetch").Err(err).Msg("fetch read error")
		return "", false, err
	}

	first := doc.Find(f.query).First()
	if first == nil {
		log.Warn().Str("stage", "fetch").Msg("target not found")
		return "", changed, nil
	}

	return first.Text(), changed, nil
}
