package youtube

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"music-bot-v2/internal/music"
)

const (
	searchCallbackPrefix     = "yt:"
	paginationCallbackPrefix = "ytp:"
	searchPageLimit          = 10
	maxButtonLabelRunes      = 64
)

func (h *Handler) searchText() handlers.Response {
	return func(b *gotgbot.Bot, ctx *ext.Context) error {
		if h == nil || h.music == nil {
			return errors.New("music consumer is nil")
		}
		if ctx == nil || ctx.EffectiveMessage == nil || ctx.EffectiveChat == nil {
			return errors.New("missing message context")
		}

		requester := requesterID(ctx)
		query := strings.TrimSpace(ctx.EffectiveMessage.GetText())
		if query == "" {
			_, err := b.SendMessageWithContext(h.ctx, ctx.EffectiveChat.Id, "Search query is empty.", nil)
			return err
		}

		go h.setQuery(requester, query)
		h.music.ResetSearchState(requester)

		items, total, err := h.music.SearchVideos(h.ctx, query, 0, requester)
		if err != nil {
			go h.clearPanelMessage(requester)
			_, sendErr := b.SendMessageWithContext(h.ctx, ctx.EffectiveChat.Id, "Search failed. Please try again later.", nil)
			if sendErr != nil {
				return sendErr
			}
			return err
		}

		if len(items) == 0 {
			go h.clearPanelMessage(requester)
			_, err = b.SendMessageWithContext(h.ctx, ctx.EffectiveChat.Id, "No videos found.", nil)
			return err
		}

		keyboard := buildSearchKeyboard(items, 0, total)
		if chatID, messageID, ok := h.getPanelMessage(requester); ok {
			_, _ = b.DeleteMessageWithContext(h.ctx, chatID, messageID, nil)
		}
		message, err := b.SendMessageWithContext(h.ctx, ctx.EffectiveChat.Id, searchMessageText(0, total), &gotgbot.SendMessageOpts{
			ReplyMarkup: keyboard,
		})
		if err == nil && message != nil {
			go h.setPanelMessage(requester, message.Chat.Id, message.MessageId)
		}
		return err
	}
}

func buildSearchKeyboard(items []music.VideoInfo, page int, total int) gotgbot.InlineKeyboardMarkup {
	rows := make([][]gotgbot.InlineKeyboardButton, 0, len(items)+1)
	for i, item := range items {
		label := fmt.Sprintf("%d. %s", i+1, item.Title)
		rows = append(rows, []gotgbot.InlineKeyboardButton{{
			Text:         trimButtonLabel(label),
			CallbackData: searchCallbackPrefix + item.ID,
		}})
	}
	if nav := buildPaginationRow(page, total); len(nav) > 0 {
		rows = append(rows, nav)
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func buildPaginationRow(page int, total int) []gotgbot.InlineKeyboardButton {
	totalPages := pageCount(total, searchPageLimit)
	if totalPages <= 1 {
		return nil
	}

	hasPrev := page > 0
	hasNext := page+1 < totalPages
	if !hasPrev && !hasNext {
		return nil
	}

	row := make([]gotgbot.InlineKeyboardButton, 0, 2)
	if hasPrev {
		row = append(row, gotgbot.InlineKeyboardButton{
			Text:         "⬅️",
			CallbackData: paginationCallbackPrefix + strconv.Itoa(page-1),
		})
	}
	if hasNext {
		row = append(row, gotgbot.InlineKeyboardButton{
			Text:         "➡️",
			CallbackData: paginationCallbackPrefix + strconv.Itoa(page+1),
		})
	}
	return row
}

func searchMessageText(page int, total int) string {
	totalPages := pageCount(total, searchPageLimit)
	if totalPages <= 1 {
		return "Select a track:"
	}
	return fmt.Sprintf("Select a track (page %d/%d):", page+1, totalPages)
}

func pageCount(total int, limit int) int {
	if total <= 0 || limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}

func trimButtonLabel(label string) string {
	if label == "" {
		return label
	}
	runes := []rune(label)
	if len(runes) <= maxButtonLabelRunes {
		return label
	}
	return strings.TrimSpace(string(runes[:maxButtonLabelRunes-3])) + "..."
}

func requesterID(ctx *ext.Context) string {
	if ctx == nil {
		return ""
	}
	if ctx.EffectiveUser != nil {
		return strconv.FormatInt(ctx.EffectiveUser.Id, 10)
	}
	if ctx.EffectiveChat != nil {
		return strconv.FormatInt(ctx.EffectiveChat.Id, 10)
	}
	return ""
}
