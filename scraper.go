package main

import "net/url"

type scraper interface {
	getWatchedField() (target string, changed bool, err error)
	url() *url.URL
	query() string
}
