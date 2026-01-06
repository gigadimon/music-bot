package youtube

import (
	"context"

	"music-bot-v2/internal/cacher"
	"music-bot-v2/internal/music"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

type musicSearcher interface {
	SearchVideos(ctx context.Context, query string, page int, requester string) ([]music.VideoInfo, int, error)
	ResetSearchState(requester string)
	MP3Link(ctx context.Context, id string) (string, error)
}

type cacherService interface {
	Get(key string) (string, bool)
	Set(key, value string)
	DeletePrefix(prefix string)
}

type Handler struct {
	ctx        context.Context
	music      musicSearcher
	cache      cacherService
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
		cache:      cacher.NewInMem(),
		panelCache: cacher.NewInMem(),
		audioCache: cacher.NewInMem(),
	}
}

func (h *Handler) Handlers() []ext.Handler {
	return []ext.Handler{
		handlers.NewMessage(message.Text, h.searchText()),
		handlers.NewCallback(callbackquery.Prefix(paginationCallbackPrefix), h.paginationCallback()),
		handlers.NewCallback(callbackquery.Prefix(searchCallbackPrefix), h.getAudioCallback()),
	}
}
