package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"music-bot-v2/internal/application/config"
	"music-bot-v2/internal/application/logger"
)

type Bot struct {
	cfg        config.Config
	handlers   []ext.Handler
	updater    *ext.Updater
	dispatcher *ext.Dispatcher
}

func New(cfg config.Config, handlers []ext.Handler) (*Bot, error) {
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Processor: logger.NewUpdateProcessor(ext.BaseProcessor{}),
		// If an error is returned by a handler, log it and continue going.
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{})

	return &Bot{
		cfg:        cfg,
		handlers:   handlers,
		updater:    updater,
		dispatcher: dispatcher,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	gtgBot, err := gotgbot.NewBot(b.cfg.BotAPIToken, nil)
	if err != nil {
		return err
	}

	for _, handler := range b.handlers {
		b.dispatcher.AddHandler(handler)
	}

	webhookURL := strings.TrimSpace(b.cfg.WebhookURL)
	if webhookURL == "" {
		return fmt.Errorf("webhook url is empty")
	}

	listenAddr := strings.TrimSpace(b.cfg.WebhookListenAddr)
	if listenAddr == "" {
		return fmt.Errorf("webhook listen addr is empty")
	}

	urlPath, err := normalizeWebhookPath(b.cfg.WebhookPath)
	if err != nil {
		return err
	}

	if err := b.updater.StartWebhook(gtgBot, urlPath, ext.WebhookOpts{
		ListenAddr:  listenAddr,
		SecretToken: b.cfg.WebhookSecretToken,
	}); err != nil {
		return err
	}

	if err := b.updater.SetAllBotWebhooks(webhookURL, &gotgbot.SetWebhookOpts{
		DropPendingUpdates: b.cfg.DropPendingUpdates,
		SecretToken:        b.cfg.WebhookSecretToken,
		RequestOpts: &gotgbot.RequestOpts{
			Timeout: time.Second * time.Duration(b.cfg.RequestTimeoutSec),
		},
	}); err != nil {
		return err
	}

	log.Printf(
		"bot started: listen_addr=%s webhook_url=%s webhook_path=%s drop_pending_updates=%t",
		listenAddr,
		webhookURL,
		urlPath,
		b.cfg.DropPendingUpdates,
	)

	stopErr := make(chan error, 1)
	var stopOnce sync.Once
	stop := func() {
		stopOnce.Do(func() {
			stopErr <- b.updater.Stop()
		})
	}
	go func() {
		<-ctx.Done()
		stop()
	}()

	b.updater.Idle()
	stop()
	return <-stopErr
}

func normalizeWebhookPath(path string) (string, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", fmt.Errorf("webhook path is empty")
	}
	if !strings.HasPrefix(p, "/") {
		return "/" + p, nil
	}
	return p, nil
}
