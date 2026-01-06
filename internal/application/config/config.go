package config

import (
	"music-bot-v2/internal/cacher"

	"github.com/caarlos0/env/v9"
)

type Config struct {
	BotAPIToken        string        `env:"CONFIGURATION_BOT_API_TOKEN"`
	DropPendingUpdates bool          `env:"CONFIGURATION_BOT_DROP_PENDING_UPDATES" envDefault:"true"`
	RequestTimeoutSec  int           `env:"CONFIGURATION_BOT_REQUEST_TIMEOUT_SEC" envDefault:"10"`
	WebhookURL         string        `env:"CONFIGURATION_BOT_WEBHOOK_URL"`
	WebhookPath        string        `env:"CONFIGURATION_BOT_WEBHOOK_PATH" envDefault:"/bot"`
	WebhookListenAddr  string        `env:"CONFIGURATION_BOT_WEBHOOK_LISTEN_ADDR" envDefault:":8080"`
	WebhookSecretToken string        `env:"CONFIGURATION_BOT_WEBHOOK_SECRET_TOKEN"`
	GoogleAPIKeys      []string      `env:"CONFIGURATION_GOOGLE_API_KEY" envSeparator:","`
	Cacher             cacher.Config `envPrefix:"CONFIGURATION_CACHER_"`
}

func Get() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
