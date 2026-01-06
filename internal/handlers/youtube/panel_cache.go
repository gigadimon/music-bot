package youtube

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Кеш сообщения панели поиска (chatID/messageID) по requester.
func (h *Handler) setPanelMessage(requester string, chatID int64, messageID int64) {
	if requester == "" {
		return
	}
	value := fmt.Sprintf("%d:%d", chatID, messageID)
	if err := h.panelCache.Set(h.ctx, requester, value); err != nil {
		log.Printf("cache set panel requester=%s err=%v", requester, err)
	}
}

func (h *Handler) getPanelMessage(requester string) (int64, int64, bool) {
	if requester == "" {
		return 0, 0, false
	}
	value, ok, err := h.panelCache.Get(h.ctx, requester)
	if err != nil {
		log.Printf("cache get panel requester=%s err=%v", requester, err)
		return 0, 0, false
	}
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
	if err := h.panelCache.Set(h.ctx, requester, ""); err != nil {
		log.Printf("cache clear panel requester=%s err=%v", requester, err)
	}
}
