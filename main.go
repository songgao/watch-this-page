package main

import (
	"flag"
	"time"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/rs/zerolog/log"
)

var url = flag.String("url", "", "url to scrape")
var query = flag.String("query", "", "a jquery-style query to locate the target element")
var interval = flag.Duration("interval", time.Minute, "scrape interval")
var convo = flag.String("convo", "", "keybase conversation name of the keybase chat to send updates to. mutually exclusive with -team")
var team = flag.String("team", "", "keybase team name of the keybase chat to send updates to. mutually exclusive with -convo")
var channel = flag.String("channel", "", "keybase chat channel name under -name of the keybase chat to send updates to. ignored unless -team is usesd")
var mention = flag.String("mention", "", "username to at-mention on target value changes")

func setupLogging() {
	log.Logger = log.With().Caller().Logger()
}

func checkArgs() {
	flag.Parse()
	if len(*url) == 0 {
		log.Fatal().Msg("missing url")
	}
	if len(*query) == 0 {
		log.Fatal().Msg("missing query")
	}
	if len(*team) == 0 && len(*convo) == 0 {
		log.Fatal().Msg("missing team or convo")
	}

	log.Info().
		Str("url", *url).
		Str("query", *query).
		Dur("interval", *interval).
		Str("convo", *convo).
		Str("team", *team).
		Str("channel", *channel).
		Str("mention", *mention).
		Msg("args")
}

func setupKeybase() (sender sender) {
	kbc, err := kbchat.Start(kbchat.RunOptions{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup keybase")
	}

	if len(*convo) != 0 {
		return func(msg string, useMention bool) error {
			if useMention && len(*mention) != 0 {
				msg = "@" + *mention + " " + msg
			}
			if _, err = kbc.SendMessageByTlfName(*convo, msg); err != nil {
				return err
			}
			return nil
		}
	} else if len(*team) != 0 {
		useChannel := (*string)(nil)
		if len(*channel) != 0 {
			useChannel = channel
		}
		return func(msg string, useMention bool) error {
			if useMention && len(*mention) != 0 {
				msg = "@" + *mention + " " + msg
			}
			if _, err = kbc.SendMessageByTeamName(*team, useChannel, msg); err != nil {
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
	run(sender, &watchedField{url: *url, query: *query})
}
