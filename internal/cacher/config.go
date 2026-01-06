package cacher

import "sync"

var (
	currentConfig = defaultConfig()
	configMu      sync.RWMutex
)

// Config describes Redis cache settings.
type Config struct {
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisUsername string `env:"REDIS_USERNAME" envDefault:"app"`
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:"local-redis-pass"`
}

func defaultConfig() Config {
	return Config{
		RedisAddr:     "localhost:6379",
		RedisUsername: "app",
		RedisPassword: "local-redis-pass",
	}
}

// SetConfig updates the global cache configuration.
func SetConfig(cfg Config) {
	configMu.Lock()
	currentConfig = cfg
	configMu.Unlock()
}

func getConfig() Config {
	configMu.RLock()
	cfg := currentConfig
	configMu.RUnlock()
	defaults := defaultConfig()
	if cfg.RedisAddr == "" {
		cfg.RedisAddr = defaults.RedisAddr
	}
	if cfg.RedisUsername == "" {
		cfg.RedisUsername = defaults.RedisUsername
	}
	return cfg
}
