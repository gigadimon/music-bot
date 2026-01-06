package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"music-bot-v2/internal/yt1s"

	"music-bot-v2/internal/music"
	"music-bot-v2/internal/youtube"

	"music-bot-v2/internal/application/bot"
	"music-bot-v2/internal/application/config"
	ytHandlers "music-bot-v2/internal/handlers/youtube"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Get()
	if err != nil {
		log.Panicln("failed to load config: " + err.Error())
	}

	ytCl := youtube.NewClient(cfg.GoogleAPIKey, nil)
	ytExtrCl := yt1s.NewClient(nil)

	ms := music.NewService(ytCl, ytExtrCl)

	h := ytHandlers.NewHandler(ctx, ms)

	b, err := bot.New(cfg, h.Handlers())
	if err != nil {
		log.Panicln("failed to create bot: " + err.Error())
	}

	if err := b.Start(ctx); err != nil {
		log.Panicln("failed to start bot:", err.Error())
	}
}
