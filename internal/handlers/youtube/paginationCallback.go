package youtube

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func (h *Handler) paginationCallback() handlers.Response {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		if h == nil || h.music == nil {
			return errors.New("music consumer is nil")
		}
		if ctx == nil || ctx.CallbackQuery == nil {
			return errors.New("missing callback query")
		}

		page, err := parsePaginationPage(ctx.CallbackQuery.Data)
		if err != nil {
			_ = answerCallback(h.ctx, b, ctx.CallbackQuery, "Invalid page.")
			return err
		}

		requester := requesterID(ctx)
		query := strings.TrimSpace(h.getQuery(requester))
		if query == "" {
			return answerCallback(h.ctx, b, ctx.CallbackQuery, "Search expired. Send a new query.")
		}

		items, total, err := h.music.SearchVideos(h.ctx, query, page, requester)
		if err != nil {
			_ = answerCallback(h.ctx, b, ctx.CallbackQuery, "Search failed. Please try again.")
			return err
		}

		if len(items) == 0 {
			return answerCallback(h.ctx, b, ctx.CallbackQuery, "No videos found.")
		}

		totalPages := pageCount(total, searchPageLimit)
		if totalPages > 0 && page >= totalPages {
			return answerCallback(h.ctx, b, ctx.CallbackQuery, "No more pages.")
		}

		if ctx.EffectiveMessage == nil {
			return errors.New("missing message to edit")
		}

		keyboard := buildSearchKeyboard(items, page, total)
		_, _, err = b.EditMessageTextWithContext(h.ctx, searchMessageText(page, total), &gotgbot.EditMessageTextOpts{
			ChatId:      ctx.EffectiveMessage.Chat.Id,
			MessageId:   ctx.EffectiveMessage.MessageId,
			ReplyMarkup: keyboard,
		})
		if err != nil {
			return err
		}

		return answerCallback(h.ctx, b, ctx.CallbackQuery, "")
	}
}

func parsePaginationPage(data string) (int, error) {
	if !strings.HasPrefix(data, paginationCallbackPrefix) {
		return 0, errors.New("unexpected callback data")
	}

	pageStr := strings.TrimPrefix(data, paginationCallbackPrefix)
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		return 0, errors.New("invalid page")
	}

	return page, nil
}

func answerCallback(ctx context.Context, b *gotgbot.Bot, cq *gotgbot.CallbackQuery, text string) error {
	if b == nil || cq == nil {
		return nil
	}
	if text == "" {
		_, err := b.AnswerCallbackQueryWithContext(ctx, cq.Id, nil)
		return err
	}
	_, err := b.AnswerCallbackQueryWithContext(ctx, cq.Id, &gotgbot.AnswerCallbackQueryOpts{Text: text})
	return err
}
