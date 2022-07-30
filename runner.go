package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type sender = func(msg string, useMention bool) error

func makeMsg(s scraper, target string) string {
	if len(target) == 0 {
		target = "<empty>"
	}
	return fmt.Sprintf("url: %s\nquery: `%s`\ncurrent value: `%s`",
		s.url().String(), s.query(), target)
}

func makeMentionMessageIfNeeded(target string, last *string) string {
	if last != nil && *last != target {
		return fmt.Sprintf("page changed! `%s` -> `%s`", *last, target)
	}
	return ""
}

func makeStartCh() <-chan time.Time {
	ret := make(chan time.Time, 1)
	ret <- time.Now()
	return ret
}

func run(ctx context.Context, sender sender, s scraper) {
	log.Debug().Msg("+ run")
	defer log.Debug().Msg("- run")

	err := sender(":rocket::rocket::rocket::rocket::rocket:\n\nI'm alive!\n\n:rocket::rocket::rocket::rocket::rocket:", false)
	if err != nil {
		log.Error().Err(err).Msg("send error")
	}

	ticker := time.Tick(*fInterval)
	startCh := makeStartCh()
	last := (*string)(nil)
	for {
		select {
		case <-ctx.Done():
			return
		case <-startCh:
		case <-ticker:
		}

		log.Debug().Msg("tick")

		target, err := s.getWatchedField()
		if err != nil {
			log.Error().Err(err).Msg("scraper error")
			err = sender(fmt.Sprintf("scraper error: %v", err), false)
			continue
		}

		log.Info().Str("target", target).Msg("current value")

		msg := makeMsg(s, target)
		err = sender(msg, false)
		if err != nil {
			log.Error().Err(err).Msg("send error")
			continue
		}

		mentionMessage := makeMentionMessageIfNeeded(target, last)
		if len(mentionMessage) > 0 {
			err = sender(mentionMessage, true)
			if err != nil {
				log.Error().Err(err).Msg("send error")
			}
		}

		last = &target
	}
}
