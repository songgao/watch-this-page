package main

import (
	"fmt"
	"net/url"
	"sync/atomic"

	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog/log"
)

const numBrowserContexts = 8
const pageTimeout = float64(4000) // 4s

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
		log.Error().Err(err).Msg("cannot launch browser")
		return nil, err
	}

	defer func() {
		if err != nil {
			log.Info().Msg("error newPlaywrightScraper. Closing the browser.")
			s.browser.Close()
		}
	}()

	return s, nil
}

func (s *playwrightScraper) shutdown() (err error) {
	log.Debug().Msg("+ playwrightScraper.shutdown()")
	defer log.Debug().Msg("- playwrightScraper.shutdown()")

	return s.browser.Close()
}

func (s *playwrightScraper) ensurePageIndex(i uint64) (err error) {
	if s.pages[i] != nil {
		return nil
	}
	page, err := s.browser.NewPage()
	if err != nil {
		log.Error().Uint64("page_index", i).Err(err).Msg("cannot create a new page")
		return err
	}
	page.SetDefaultNavigationTimeout(pageTimeout) // 4s
	page.SetDefaultTimeout(pageTimeout)           // 4s
	if _, err = page.Goto(s.u.String()); err != nil {
		log.Error().Uint64("page_index", i).Err(err).Msg("cannot go to the URL")
		return err
	}
	s.pages[i] = page
	return nil
}

func (s *playwrightScraper) getWatchedField() (target string, err error) {
	pageIndex := atomic.AddUint64(&s.callCount, 1) % numBrowserContexts

	log.Debug().Uint64("page_index", pageIndex).Msg("+ getWatchedField")
	defer log.Debug().Uint64("page_index", pageIndex).Msg("- getWatchedField")

	if err = s.ensurePageIndex(pageIndex); err != nil {
		log.Error().Err(err).Msg("ensure page error")
		return "", err
	}

	page := s.pages[pageIndex]
	if _, err = page.Reload(); err != nil {
		log.Error().Err(err).Msg("cannot reload")
		return "", err
	}
	entries, err := page.QuerySelectorAll(s.q)
	if err != nil {
		log.Error().Err(err).Msg("cannot get entries")
		return "", err
	}
	if len(entries) != 1 {
		log.Error().Int("num_entries", len(entries)).Msg("unexpected number of entries")
		return "", fmt.Errorf("unexpected number of entries %v", len(entries))
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
