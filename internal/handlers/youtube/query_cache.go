package youtube

import (
	"log"
)

// Кеш поискового запроса по requester используемый при переходе между страницами нажатием на кнопки.
func (h *Handler) setQuery(requester string, query string) {
	if requester == "" {
		return
	}
	if err := h.queryCache.Set(h.ctx, requester, query); err != nil {
		log.Printf("cache set query requester=%s err=%v", requester, err)
	}
}

func (h *Handler) getQuery(requester string) string {
	if requester == "" {
		return ""
	}
	query, ok, err := h.queryCache.Get(h.ctx, requester)
	if err != nil {
		log.Printf("cache get query requester=%s err=%v", requester, err)
		return ""
	}
	if !ok {
		return ""
	}
	return query
}
