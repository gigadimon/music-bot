package youtube

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	}
}

func (h *Handler) Handlers() []ext.Handler {
	return []ext.Handler{
		handlers.NewMessage(message.Text, h.searchText()),
		handlers.NewCallback(callbackquery.Prefix(paginationCallbackPrefix), h.paginationCallback()),
		handlers.NewCallback(callbackquery.Prefix(searchCallbackPrefix), h.getAudioCallback()),
	}
}

func (h *Handler) setQuery(requester string, query string) {
	if requester == "" {
		return
	}
	h.cache.Set(requester, query)
}

func (h *Handler) getQuery(requester string) string {
	if requester == "" {
		return ""
	}
	query, _ := h.cache.Get(requester)
	return query
}

func (h *Handler) setPanelMessage(requester string, chatID int64, messageID int64) {
	if requester == "" {
		return
	}
	value := fmt.Sprintf("%d:%d", chatID, messageID)
	h.panelCache.Set(requester, value)
}

func (h *Handler) getPanelMessage(requester string) (int64, int64, bool) {
	if requester == "" {
		return 0, 0, false
	}
	value, ok := h.panelCache.Get(requester)
	if !ok || value == "" {
		return 0, 0, false
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	chatID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	messageID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return chatID, messageID, true
}

func (h *Handler) clearPanelMessage(requester string) {
	if requester == "" {
		return
	}
	h.panelCache.Set(requester, "")
}
