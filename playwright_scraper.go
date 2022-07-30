package main

import (
	"errors"
	"net/url"
)

type playwrightScraper struct {
	u *url.URL
	q string
}

var _ scraper = (*playwrightScraper)(nil)

func newPlaywrightScraper(pageURL string, query string) (*playwrightScraper, error) {
	u, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}

	return &playwrightScraper{
		u: u,
		q: query,
	}, nil
}

func (s *playwrightScraper) getWatchedField() (target string, changed bool, err error) {
	return "", false, errors.New("not implemented")
}

func (s *playwrightScraper) url() *url.URL {
	return s.u
}

func (s *playwrightScraper) query() string {
	return s.q
}
