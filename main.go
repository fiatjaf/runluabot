package main

import (
	"os"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

type Settings struct {
	BotToken string `envconfig:"BOT_TOKEN" required:"true"`
}

var err error
var s Settings
var bot *tgbotapi.BotAPI
var log = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr})
var router = mux.NewRouter()

func main() {
	err = envconfig.Process("", &s)
	if err != nil {
		log.Fatal().Err(err).Msg("couldn't process envconfig.")
	}

	// bot stuff
	bot, err = tgbotapi.NewBotAPI(s.BotToken)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	log.Info().Str("username", bot.Self.UserName).Msg("telegram bot authorized")

	var lastTelegramUpdate int64 = -1

	u := tgbotapi.NewUpdate(int(lastTelegramUpdate + 1))
	u.Timeout = 600
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Error().Err(err).Msg("telegram getupdates fail")
		return
	}

	for update := range updates {
		lastTelegramUpdate = int64(update.UpdateID)
		handle(update)
	}
}
