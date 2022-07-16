package main

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type sender = func(msg string, useMention bool) error

func run(sender sender, watchedField *watchedField) {
	err := sender(":rocket::rocket::rocket::rocket::rocket:\n\nI'm alive!\n\n:rocket::rocket::rocket::rocket::rocket:", false)
	if err != nil {
		log.Error().Err(err).Msg("send error")
	}

	ticker := time.Tick(*interval)
	last := (*string)(nil)
	for range ticker {
		log.Debug().Msg("tick")

		target, err := getWatchedField(watchedField)
		if err != nil {
			log.Error().Err(err).Msg("fetch error")
			continue
		}

		log.Info().Str("target", target).Msg("current value")

		err = sender(
			fmt.Sprintf("url: %s\nquery: `%s`\ncurrent value: `%s`",
				watchedField.url, watchedField.query, target),
			false)
		if err != nil {
			log.Error().Err(err).Msg("send error")
			continue
		}

		if last != nil && *last != target {
			sender(
				fmt.Sprintf("value changed! `%s` -> `%s`", last, target),
				true)
		}
		if err != nil {
			log.Error().Err(err).Msg("send error")
			continue
		}

		last = &target
	}
}
