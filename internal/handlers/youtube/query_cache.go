package youtube

// Кеш поискового запроса по requester используемый при переходе между страницами нажатием на кнопки.
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
