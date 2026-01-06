package youtube

import "log"

// Кеш file_id аудио по trackID.
func (h *Handler) setAudioFileID(trackID string, fileID string) {
	if trackID == "" || fileID == "" {
		return
	}
	if err := h.audioCache.Set(h.ctx, trackID, fileID); err != nil {
		log.Printf("cache set audio track_id=%s err=%v", trackID, err)
	}
}

func (h *Handler) getAudioFileID(trackID string) string {
	if trackID == "" {
		return ""
	}
	fileID, ok, err := h.audioCache.Get(h.ctx, trackID)
	if err != nil {
		log.Printf("cache get audio track_id=%s err=%v", trackID, err)
		return ""
	}
	if !ok || fileID == "" {
		return ""
	}
	return fileID
}

func (h *Handler) clearAudioFileID(trackID string) {
	if trackID == "" {
		return
	}
	if err := h.audioCache.Set(h.ctx, trackID, ""); err != nil {
		log.Printf("cache clear audio track_id=%s err=%v", trackID, err)
	}
}
