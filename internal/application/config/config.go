package config

import (
	"music-bot-v2/internal/cacher"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	BotAPIToken        string        `env:"BOT_API_TOKEN"`
	GoogleAPIKey       string        `env:"GOOGLE_API_KEY"`
	DropPendingUpdates bool          `env:"BOT_DROP_PENDING_UPDATES" envDefault:"true"`
	RequestTimeoutSec  int           `env:"BOT_REQUEST_TIMEOUT_SEC" envDefault:"10"`
	WebhookURL         string        `env:"BOT_WEBHOOK_URL"`
	WebhookPath        string        `env:"BOT_WEBHOOK_PATH" envDefault:"/bot"`
	WebhookListenAddr  string        `env:"BOT_WEBHOOK_LISTEN_ADDR" envDefault:":8080"`
	WebhookSecretToken string        `env:"BOT_WEBHOOK_SECRET_TOKEN"`
	Cacher             cacher.Config `envPrefix:"CACHER_"`
}

func Get() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
