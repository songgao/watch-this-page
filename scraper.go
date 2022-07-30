package main

import "net/url"

type scraper interface {
	getWatchedField() (target string, err error)
	shutdown() error

	url() *url.URL
	query() string
}
