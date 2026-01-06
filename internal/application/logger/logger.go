package logger

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type updateProcessor struct {
	next ext.Processor
}

func NewUpdateProcessor(next ext.Processor) ext.Processor {
	return updateProcessor{next: next}
}

func (up updateProcessor) ProcessUpdate(d *ext.Dispatcher, b *gotgbot.Bot, ctx *ext.Context) error {
	start := time.Now()
	if up.next == nil {
		logUpdate(ctx, time.Since(start))
		return nil
	}
	err := up.next.ProcessUpdate(d, b, ctx)
	logUpdate(ctx, time.Since(start))
	return err
}

func logUpdate(ctx *ext.Context, elapsed time.Duration) {
	elapsedMs := elapsed.Milliseconds()
	if ctx == nil || ctx.Update == nil {
		log.Printf("update received: <nil> duration_ms=%d", elapsedMs)
		return
	}

	updateType, details := describeUpdate(ctx)
	updateID := ctx.Update.UpdateId

	var chatID int64
	if ctx.EffectiveChat != nil {
		chatID = ctx.EffectiveChat.Id
	}

	var userID int64
	var username string
	if ctx.EffectiveUser != nil {
		userID = ctx.EffectiveUser.Id
		username = ctx.EffectiveUser.Username
	}

	if username == "" {
		username = "-"
	}

	if details == "" {
		log.Printf("update id=%d type=%s chat_id=%d user_id=%d username=%s duration_ms=%d", updateID, updateType, chatID, userID, username, elapsedMs)
		return
	}

	log.Printf("update id=%d type=%s chat_id=%d user_id=%d username=%s %s duration_ms=%d", updateID, updateType, chatID, userID, username, details, elapsedMs)
}

func describeUpdate(ctx *ext.Context) (string, string) {
	switch {
	case ctx.Message != nil:
		text := strings.TrimSpace(ctx.Message.Text)
		return "message", messageDetails(text, int(ctx.Message.MessageId))
	case ctx.EditedMessage != nil:
		text := strings.TrimSpace(ctx.EditedMessage.Text)
		return "edited_message", messageDetails(text, int(ctx.EditedMessage.MessageId))
	case ctx.ChannelPost != nil:
		text := strings.TrimSpace(ctx.ChannelPost.Text)
		return "channel_post", messageDetails(text, int(ctx.ChannelPost.MessageId))
	case ctx.EditedChannelPost != nil:
		text := strings.TrimSpace(ctx.EditedChannelPost.Text)
		return "edited_channel_post", messageDetails(text, int(ctx.EditedChannelPost.MessageId))
	case ctx.CallbackQuery != nil:
		data := strings.TrimSpace(ctx.CallbackQuery.Data)
		return "callback_query", callbackDetails(data, ctx.CallbackQuery.Id)
	case ctx.InlineQuery != nil:
		query := strings.TrimSpace(ctx.InlineQuery.Query)
		return "inline_query", inlineDetails(query, ctx.InlineQuery.Id)
	case ctx.ChosenInlineResult != nil:
		resultID := strings.TrimSpace(ctx.ChosenInlineResult.ResultId)
		return "chosen_inline_result", chosenInlineDetails(resultID)
	case ctx.ShippingQuery != nil:
		return "shipping_query", ""
	case ctx.PreCheckoutQuery != nil:
		return "pre_checkout_query", ""
	case ctx.Poll != nil:
		return "poll", ""
	case ctx.PollAnswer != nil:
		return "poll_answer", ""
	case ctx.MyChatMember != nil:
		return "my_chat_member", ""
	case ctx.ChatMember != nil:
		return "chat_member", ""
	case ctx.ChatJoinRequest != nil:
		return "chat_join_request", ""
	case ctx.BusinessConnection != nil:
		return "business_connection", ""
	case ctx.BusinessMessage != nil:
		text := strings.TrimSpace(ctx.BusinessMessage.Text)
		return "business_message", messageDetails(text, int(ctx.BusinessMessage.MessageId))
	case ctx.EditedBusinessMessage != nil:
		text := strings.TrimSpace(ctx.EditedBusinessMessage.Text)
		return "edited_business_message", messageDetails(text, int(ctx.EditedBusinessMessage.MessageId))
	case ctx.DeletedBusinessMessages != nil:
		return "deleted_business_messages", ""
	case ctx.ChatBoost != nil:
		return "chat_boost", ""
	case ctx.RemovedChatBoost != nil:
		return "removed_chat_boost", ""
	default:
		return "unknown", ""
	}
}

func messageDetails(text string, messageID int) string {
	if text == "" {
		return fmt.Sprintf("message_id=%d", messageID)
	}
	return fmt.Sprintf("text=%q", truncateForLog(text, 80))
}

func callbackDetails(data string, callbackID string) string {
	if data == "" {
		return fmt.Sprintf("callback_id=%s", callbackID)
	}
	return fmt.Sprintf("data=%q", truncateForLog(data, 80))
}

func inlineDetails(query string, inlineID string) string {
	if query == "" {
		return fmt.Sprintf("inline_id=%s", inlineID)
	}
	return fmt.Sprintf("query=%q", truncateForLog(query, 80))
}

func chosenInlineDetails(resultID string) string {
	if resultID == "" {
		return ""
	}
	return fmt.Sprintf("result_id=%s", truncateForLog(resultID, 80))
}

func truncateForLog(value string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}
