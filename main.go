package main

import (
	"net/url"

	"github.com/caarlos0/env/v10"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type AllowList []int64

func (s AllowList) Allowed(id int64) bool {
	if len(s) == 0 {
		return true
	}
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}

type Config struct {
	TelegramToken     string    `env:"TG_TOKEN,notEmpty"`
	TelegramAllowList AllowList `env:"TG_ALLOWLIST"`

	WallabagURL          string `env:"WB_URL,notEmpty"`
	WallabagClientID     string `env:"WB_CLIENT_ID,notEmpty"`
	WallabagClientSecret string `env:"WB_CLIENT_SECRET,notEmpty,unset"`
	WallabagUsername     string `env:"WB_USERNAME,notEmpty"`
	WallabagPassword     string `env:"WB_PASSWORD,notEmpty,unset"`
}

func main() {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.Fatal("Error parsing config: ", err)
	}

	wb := Wallabag{
		Url:          cfg.WallabagURL,
		ClientID:     cfg.WallabagClientID,
		ClientSecret: cfg.WallabagClientSecret,
		Username:     cfg.WallabagUsername,
		Password:     cfg.WallabagPassword,
	}
	if err := wb.Test(); err != nil {
		logrus.Fatalf("Unable to connect to wallbag: %s", err)
	}

	bot, err := tgbot.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		logrus.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	logrus.Info("Bot is listening...")
	for update := range bot.GetUpdatesChan(u) {
		var (
			fromId = update.Message.From.ID
			chatId = update.Message.Chat.ID
			text   = update.Message.Text
		)

		logrus.Infof("%d (%d): %s", fromId, chatId, text)

		if !cfg.TelegramAllowList.Allowed(fromId) {
			bot.Send(tgbotapi.NewMessage(chatId, "This user is not allowed to communicate to this bot"))
			logrus.Warnf("%d not on allowlist", fromId)
			continue
		}

		if _, err := url.Parse(text); err != nil {
			bot.Send(tgbotapi.NewMessage(chatId, "Bad URL"))
			continue
		}

		if err := wb.AddURL(text); err != nil {
			logrus.Errorf("Error posting URL: %s", err)
			bot.Send(tgbotapi.NewMessage(chatId, "Error posting URL. Check logs."))
			continue
		}

		bot.Send(tgbotapi.NewMessage(chatId, "Got it!"))
	}
}
