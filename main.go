package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/rs/zerolog/log"
)

var fURL = flag.String("url", "", "url to scrape")
var fQuery = flag.String("query", "", "a jquery-style query to locate the target element")
var fInterval = flag.Duration("interval", time.Minute, "scrape interval")
var fConvo = flag.String("convo", "", "keybase conversation name of the keybase chat to send updates to. mutually exclusive with -team")
var fTeam = flag.String("team", "", "keybase team name of the keybase chat to send updates to. mutually exclusive with -convo")
var fChannel = flag.String("channel", "", "keybase chat channel name under -name of the keybase chat to send updates to. ignored unless -team is usesd")
var fMention = flag.String("mention", "", "username to at-mention on target value changes")

func setupLogging() {
	log.Logger = log.With().Caller().Logger()
}

func checkArgs() {
	flag.Parse()
	if len(*fURL) == 0 {
		log.Fatal().Msg("missing url")
	}
	if len(*fQuery) == 0 {
		log.Fatal().Msg("missing query")
	}
	if len(*fTeam) == 0 && len(*fConvo) == 0 {
		log.Fatal().Msg("missing team or convo")
	}

	log.Info().
		Str("url", *fURL).
		Str("query", *fQuery).
		Dur("interval", *fInterval).
		Str("convo", *fConvo).
		Str("team", *fTeam).
		Str("channel", *fChannel).
		Str("mention", *fMention).
		Msg("args")
}

func setupKeybase() (sender sender) {
	kbc, err := kbchat.Start(kbchat.RunOptions{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup keybase")
	}

	if len(*fConvo) != 0 {
		return func(msg string, useMention bool) error {
			if useMention && len(*fMention) != 0 {
				msg = "@" + *fMention + " " + msg
			}
			if _, err = kbc.SendMessageByTlfName(*fConvo, msg); err != nil {
				return err
			}
			return nil
		}
	} else if len(*fTeam) != 0 {
		useChannel := (*string)(nil)
		if len(*fChannel) != 0 {
			useChannel = fChannel
		}
		return func(msg string, useMention bool) error {
			if useMention && len(*fMention) != 0 {
				msg = "@" + *fMention + " " + msg
			}
			if _, err = kbc.SendMessageByTeamName(*fTeam, useChannel, msg); err != nil {
				return err
			}
			return nil
		}
	}
	log.Fatal().Msg("programmer error")
	return nil
}

func main() {
	setupLogging()
	checkArgs()
	sender := setupKeybase()
	s, err := newPlaywrightScraper(*fURL, *fQuery)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create watcher")
	}
	defer s.shutdown()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()
	run(ctx, sender, s)
}
