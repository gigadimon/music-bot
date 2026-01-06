package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/labstack/echo/v4"

	"music-bot-v2/internal/application/config"
	"music-bot-v2/internal/application/logger"
)

type Mountable interface {
	Mount(e *echo.Echo)
}

type Bot struct {
	cfg        config.Config
	handlers   []ext.Handler
	updater    *ext.Updater
	dispatcher *ext.Dispatcher
	server     *http.Server
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

func (b *Bot) Start(ctx context.Context, mountable ...Mountable) error {
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

	if err := b.updater.AddWebhook(gtgBot, urlPath, &ext.AddWebhookOpts{
		SecretToken: b.cfg.WebhookSecretToken,
	}); err != nil {
		return err
	}

	if err := b.startWebhookServer(listenAddr, urlPath, mountable...); err != nil {
		_ = b.updater.Stop()
		return err
	}

	if err := b.updater.SetAllBotWebhooks(webhookURL, &gotgbot.SetWebhookOpts{
		DropPendingUpdates: b.cfg.DropPendingUpdates,
		SecretToken:        b.cfg.WebhookSecretToken,
		RequestOpts: &gotgbot.RequestOpts{
			Timeout: time.Second * time.Duration(b.cfg.RequestTimeoutSec),
		},
	}); err != nil {
		b.shutdownWebhookServer()
		_ = b.updater.Stop()
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
			b.shutdownWebhookServer()
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

func (b *Bot) startWebhookServer(listenAddr string, urlPath string, mountable ...Mountable) error {
	if b.server != nil {
		return fmt.Errorf("webhook server already started")
	}

	e := echo.New()
	e.POST(urlPath, echo.WrapHandler(b.updater.GetHandlerFunc("/")))

	for _, m := range mountable {
		m.Mount(e)
	}

	server := &http.Server{
		Addr:    listenAddr,
		Handler: e,
	}
	b.server = server

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic("http server failed: " + err.Error())
		}
	}()

	return nil
}

func (b *Bot) shutdownWebhookServer() {
	if b.server == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := b.server.Shutdown(ctx)
	if err != nil {
		log.Printf("failed to shutdown webhook server: %v", err)
	}

	b.server = nil
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
