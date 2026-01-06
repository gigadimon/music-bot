package youtube

import (
	"fmt"
	"strconv"
	"strings"
)

// Кеш сообщения панели поиска (chatID/messageID) по requester.
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
