package main

import (
	"fmt"
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
	WallabagTags         string `env:"WB_TAGS" envDefault:"telegram"`
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
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

		bot.Send(tgbotapi.NewChatAction(chatId, tgbotapi.ChatTyping))

		if !cfg.TelegramAllowList.Allowed(fromId) {
			bot.Send(tgbotapi.NewMessage(chatId, "This user is not allowed to communicate to this bot"))
			logrus.Warnf("%d not on allowlist", fromId)
			continue
		}

		if !IsUrl(text) {
			bot.Send(tgbotapi.NewMessage(chatId, "I couldn't parse that as a URL"))
			continue
		}

		if articleId, err := wb.AddURL(text, cfg.WallabagTags); err != nil {
			logrus.Errorf("Error posting URL: %s", err)
			bot.Send(tgbotapi.NewMessage(chatId, fmt.Sprintf("Error posting URL. Check logs. %s", err.Error())))
		} else {
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("Got it: %s/view/%d", cfg.WallabagURL, articleId))
			msg.DisableNotification = true
			bot.Send(msg)
		}
	}
}
