package main

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type sender = func(msg string, useMention bool) error

func makeMsg(w scraper, target string, last *string) string {
	if len(target) == 0 && last != nil {
		target = *last
	}
	if len(target) == 0 {
		target = "<empty>"
	}
	return fmt.Sprintf("url: %s\nquery: `%s`\ncurrent value: `%s`",
		w.url().String(), w.query(), target)
}

func run(sender sender, w scraper) {
	err := sender(":rocket::rocket::rocket::rocket::rocket:\n\nI'm alive!\n\n:rocket::rocket::rocket::rocket::rocket:", false)
	if err != nil {
		log.Error().Err(err).Msg("send error")
	}

	ticker := time.Tick(*fInterval)
	last := (*string)(nil)
	for range ticker {
		log.Debug().Msg("tick")

		target, changed, err := w.getWatchedField()
		if err != nil {
			log.Error().Err(err).Msg("fetch error")
			err = sender(fmt.Sprintf("fetch error: %v", err), false)
			continue
		}

		log.Info().
			Str("target", target).
			Bool("changed", changed).
			Msg("current value")

		msg := makeMsg(w, target, last)
		err = sender(msg, false)
		if err != nil {
			log.Error().Err(err).Msg("send error")
			continue
		}

		if changed {
			if last != nil && *last != target {
				if err = sender(
					fmt.Sprintf("page changed! `%s` -> `%s`", *last, target),
					true); err != nil {
					log.Error().Err(err).Msg("send error")
					continue
				}
			}
			// only set if chanted==true. Otherwise target is empty
			// TODO: track this in the scraper
			last = &target
		}

	}
}
