# Wallabag telegram bot

## Config
Expects environmental variables

```
TelegramToken     string    `env:"TG_TOKEN,notEmpty"`
TelegramAllowList AllowList `env:"TG_ALLOWLIST"`

WallabagURL          string `env:"WB_URL,notEmpty"`
WallabagClientID     string `env:"WB_CLIENT_ID,notEmpty"`
WallabagClientSecret string `env:"WB_CLIENT_SECRET,notEmpty,unset"`
WallabagUsername     string `env:"WB_USERNAME,notEmpty"`
WallabagPassword     string `env:"WB_PASSWORD,notEmpty,unset"`
```

## Run
`docker build .` or `go run .`

## License
MIT
