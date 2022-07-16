package main

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)

type watchedField struct {
	url   string
	query string
}

func getWatchedField(f *watchedField) (string, error) {
	res, err := http.Get(f.url)
	if err != nil {
		log.Error().Str("stage", "fetch").Err(err).Msg("fetch error")
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Error().Str("stage", "fetch").Int("http_status_code", res.StatusCode).Str("http_status", res.Status).Msg("fetch error")
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Error().Str("stage", "fetch").Err(err).Msg("fetch error")
		return "", err
	}

	first := doc.Find(f.query).First()
	if first == nil {
		return "", nil
	}
	fmt.Printf("%#+v\n", first.Text())

	return first.Text(), nil
}
