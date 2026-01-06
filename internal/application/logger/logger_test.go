package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func TestUpdateProcessorLogsMessage(t *testing.T) {
	var buf bytes.Buffer
	restore := captureLogs(&buf)
	defer restore()

	msg := gotgbot.Message{
		MessageId: 42,
		Text:      "hello world",
		Chat:      gotgbot.Chat{Id: 1001},
		From:      &gotgbot.User{Id: 2002, Username: "alice"},
	}
	update := &gotgbot.Update{UpdateId: 3, Message: &msg}
	ctx := &ext.Context{
		Update:           update,
		EffectiveChat:    &msg.Chat,
		EffectiveUser:    msg.From,
		EffectiveMessage: &msg,
	}

	processor := NewUpdateProcessor(nil)
	if err := processor.ProcessUpdate(nil, nil, ctx); err != nil {
		t.Fatalf("ProcessUpdate failed: %v", err)
	}

	waitForLog(t, &buf, []string{
		"update id=3",
		"type=message",
		"chat_id=1001",
		"user_id=2002",
		"username=alice",
		`text="hello world"`,
		"duration_ms=",
	})
}

func TestUpdateProcessorLogsCallbackQuery(t *testing.T) {
	var buf bytes.Buffer
	restore := captureLogs(&buf)
	defer restore()

	cq := &gotgbot.CallbackQuery{
		Id:   "cb-id",
		Data: "button:next",
		From: gotgbot.User{Id: 9, Username: "bob"},
		Message: &gotgbot.Message{
			MessageId: 5,
			Chat:      gotgbot.Chat{Id: 77},
		},
	}
	update := &gotgbot.Update{UpdateId: 11, CallbackQuery: cq}
	chat := cq.Message.GetChat()
	ctx := &ext.Context{
		Update:        update,
		EffectiveChat: &chat,
		EffectiveUser: &cq.From,
	}

	processor := NewUpdateProcessor(nil)
	if err := processor.ProcessUpdate(nil, nil, ctx); err != nil {
		t.Fatalf("ProcessUpdate failed: %v", err)
	}

	waitForLog(t, &buf, []string{
		"update id=11",
		"type=callback_query",
		"chat_id=77",
		"user_id=9",
		"username=bob",
		`data="button:next"`,
		"duration_ms=",
	})
}

func TestUpdateProcessorLogsNilContext(t *testing.T) {
	var buf bytes.Buffer
	restore := captureLogs(&buf)
	defer restore()

	processor := NewUpdateProcessor(nil)
	if err := processor.ProcessUpdate(nil, nil, nil); err != nil {
		t.Fatalf("ProcessUpdate failed: %v", err)
	}

	waitForLog(t, &buf, []string{"update received: <nil>", "duration_ms="})
}

func captureLogs(buf *bytes.Buffer) func() {
	oldOut := log.Writer()
	oldFlags := log.Flags()
	log.SetOutput(buf)
	log.SetFlags(0)
	return func() {
		log.SetOutput(oldOut)
		log.SetFlags(oldFlags)
	}
}

func waitForLog(t *testing.T, buf *bytes.Buffer, substrings []string) {
	t.Helper()

	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		out := buf.String()
		if containsAll(out, substrings) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("log output missing, got: %q", out)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func containsAll(output string, substrings []string) bool {
	for _, s := range substrings {
		if !strings.Contains(output, s) {
			return false
		}
	}
	return true
}
