package youtube

// Кеш file_id аудио по trackID.
func (h *Handler) setAudioFileID(trackID string, fileID string) {
	if trackID == "" || fileID == "" {
		return
	}
	h.audioCache.Set(trackID, fileID)
}

func (h *Handler) getAudioFileID(trackID string) string {
	if trackID == "" {
		return ""
	}
	fileID, ok := h.audioCache.Get(trackID)
	if !ok || fileID == "" {
		return ""
	}
	return fileID
}

func (h *Handler) clearAudioFileID(trackID string) {
	if trackID == "" {
		return
	}
	h.audioCache.Set(trackID, "")
}
