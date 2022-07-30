package main

import (
	"net/url"
	"sync/atomic"

	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

const numBrowserContexts = 32

type playwrightScraper struct {
	u *url.URL
	q string

	browser   playwright.Browser
	pages     [numBrowserContexts]playwright.Page
	callCount uint64
}

var _ scraper = (*playwrightScraper)(nil)

func newPlaywrightScraper(pageURL string, query string) (s *playwrightScraper, err error) {
	log.Debug().Msg("+ newPlaywrightScraper")
	defer log.Debug().Msg("- newPlaywrightScraper")

	u, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}

	s = &playwrightScraper{
		u: u,
		q: query,
	}

	err = playwright.Install()
	if err != nil {
		return nil, err
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Error().Err(err).Msg("could not start playwright")
		return nil, err
	}
	s.browser, err = pw.Chromium.Launch()
	if err != nil {
		log.Error().Err(err).Msg("cannot launch chromium")
		return nil, err
	}

	defer func() {
		if err != nil {
			log.Info().Msg("error newPlaywrightScraper. Closing the browser.")
			s.browser.Close()
		}
	}()

	for i := 0; i < numBrowserContexts; i++ {
		s.pages[i], err = s.browser.NewPage()
		if err != nil {
			log.Error().Err(err).Int("page_index", i).Msg("cannot create a new page")
			return nil, err
		}
	}

	return s, nil
}

func (s *playwrightScraper) shutdown() (err error) {
	log.Debug().Msg("+ playwrightScraper.shutdown()")
	defer log.Debug().Msg("- playwrightScraper.shutdown()")

	return s.browser.Close()
}

func (s *playwrightScraper) getWatchedField() (target string, err error) {
	pageIndex := atomic.AddUint64(&s.callCount, 1)
	log.Debug().Uint64("page_index", pageIndex).Msg("+ getWatchedField")
	defer log.Debug().Uint64("page_index", pageIndex).Msg("- getWatchedField")
	page := s.pages[pageIndex]
	if _, err = page.Goto(s.u.String()); err != nil {
		log.Error().Err(err).Msg("cannot go to the URL")
		return "", err
	}
	entries, err := page.QuerySelectorAll(s.q)
	if err != nil {
		log.Error().Err(err).Msg("cannot get entries")
		return "", err
	}
	text, err := entries[0].TextContent()
	if err != nil {
		log.Error().Err(err).Msg("cannot get text content")
		return "", err
	}
	return text, nil
}

func (s *playwrightScraper) url() *url.URL {
	return s.u
}

func (s *playwrightScraper) query() string {
	return s.q
}
