package youtube

import (
	"errors"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func (h *Handler) getAudioCallback() handlers.Response {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		if h == nil || h.music == nil {
			return errors.New("music consumer is nil")
		}
		if ctx == nil || ctx.CallbackQuery == nil || ctx.EffectiveChat == nil {
			return errors.New("missing callback query context")
		}

		trackID, err := parseTrackID(ctx.CallbackQuery.Data)
		if err != nil {
			_ = answerCallback(h.ctx, b, ctx.CallbackQuery, "Invalid track.")
			return err
		}

		link, err := h.music.MP3Link(h.ctx, trackID)
		if err != nil {
			_ = answerCallback(h.ctx, b, ctx.CallbackQuery, "Failed to load track.")
			return err
		}

		_, err = b.SendAudioWithContext(h.ctx, ctx.EffectiveChat.Id, gotgbot.InputFileByURL(link), nil)
		if err != nil {
			return err
		}

		return answerCallback(h.ctx, b, ctx.CallbackQuery, "")
	}
}

func parseTrackID(data string) (string, error) {
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("invalid track id")
	}
	return parts[1], nil
}
