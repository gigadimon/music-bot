package youtube

import (
	"context"
	"time"

	"music-bot-v2/internal/cacher"
	"music-bot-v2/internal/music"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

type musicSearcher interface {
	SearchVideos(ctx context.Context, query string, page int, requester string) ([]music.VideoInfo, int, error)
	ResetSearchState(ctx context.Context, requester string)
	MP3Link(ctx context.Context, id string) (string, error)
}

type cacherService interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string) error
	DeletePrefix(ctx context.Context, prefix string) error
}

type Handler struct {
	ctx        context.Context
	music      musicSearcher
	queryCache cacherService
	panelCache cacherService
	audioCache cacherService
}

func NewHandler(ctx context.Context, music musicSearcher) *Handler {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Handler{
		ctx:        ctx,
		music:      music,
		queryCache: cacher.NewRedis(cacher.QueryCacheDB, 0),
		panelCache: cacher.NewRedis(cacher.PanelCacheDB, 48*time.Hour),
		audioCache: cacher.NewRedis(cacher.AudioCacheDB, 0),
	}
}

func (h *Handler) Handlers() []ext.Handler {
	return []ext.Handler{
		handlers.NewMessage(message.Text, h.searchText()),
		handlers.NewCallback(callbackquery.Prefix(paginationCallbackPrefix), h.paginationCallback()),
		handlers.NewCallback(callbackquery.Prefix(searchCallbackPrefix), h.getAudioCallback()),
	}
}
