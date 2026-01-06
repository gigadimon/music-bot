package music

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"music-bot-v2/internal/cacher"
	"music-bot-v2/internal/youtube"
)

type cacherService interface {
	Get(key string) (string, bool)
	Set(key, value string)
	DeletePrefix(prefix string)
}

type youtubeClient interface {
	Search(ctx context.Context, query string, pageToken string) ([]string, youtube.Pagination, error)
	Videos(ctx context.Context, ids []string) (map[string]string, error)
}

type youtubeLinkExtractorClient interface {
	MP3Link(ctx context.Context, id string) (string, error)
}

type Service struct {
	searchCache cacherService
	tokenCache  cacherService
	mp3Cache    cacherService

	youtubeClient       youtubeClient
	linkExtractorClient youtubeLinkExtractorClient
}

type VideoInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func NewService(youtubeClient youtubeClient, linkExtractorClient youtubeLinkExtractorClient) *Service {
	return &Service{
		searchCache:         cacher.NewInMem(),
		tokenCache:          cacher.NewInMem(),
		mp3Cache:            cacher.NewInMem(),
		youtubeClient:       youtubeClient,
		linkExtractorClient: linkExtractorClient,
	}
}

func (s *Service) SearchVideos(ctx context.Context, query string, page int, requester string) ([]VideoInfo, int, error) {
	if page < 0 {
		return nil, 0, errors.New("page must be non-negative")
	}

	searchKey := buildCacheKey(requester, query, strconv.Itoa(page))

	if cachedValue, ok := s.searchCache.Get(searchKey); ok {
		var cachedResult cachedSearchResult
		if err := json.Unmarshal([]byte(cachedValue), &cachedResult); err == nil {
			return cachedResult.Items, cachedResult.TotalResults, nil
		}
	}

	pageToken := ""
	if page > 0 {
		tokenKey := buildCacheKey(requester, strconv.Itoa(page))
		if cachedToken, ok := s.tokenCache.Get(tokenKey); ok {
			pageToken = cachedToken
		}
	}

	ids, pagination, err := s.youtubeClient.Search(ctx, query, pageToken)
	if err != nil {
		return nil, 0, err
	}

	titles, err := s.youtubeClient.Videos(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	items := make([]VideoInfo, 0, len(ids))
	for _, id := range ids {
		title, ok := titles[id]
		if !ok {
			continue
		}
		items = append(items, VideoInfo{
			ID:    id,
			Title: title,
		})
	}

	go s.storePageTokens(requester, page, pagination.NextPageToken, pagination.PrevPageToken)
	go s.storeSearch(searchKey, items, pagination.TotalResults)

	return items, pagination.TotalResults, nil
}

func (s *Service) MP3Link(ctx context.Context, id string) (string, error) {
	return s.linkExtractorClient.MP3Link(ctx, id)
}

func (s *Service) ResetSearchState(requester string) {
	s.tokenCache.DeletePrefix(requester)
	s.searchCache.DeletePrefix(requester)
}

func (s *Service) storePageTokens(requester string, page int, nextToken string, prevToken string) {
	if nextToken != "" {
		s.tokenCache.Set(
			buildCacheKey(requester, strconv.Itoa(page+1)),
			nextToken,
		)
	}

	if prevToken != "" {
		s.tokenCache.Set(
			buildCacheKey(requester, strconv.Itoa(page-1)),
			prevToken,
		)
	}
}

type cachedSearchResult struct {
	Items        []VideoInfo `json:"items"`
	TotalResults int         `json:"total_results"`
}

func (s *Service) storeSearch(searchKey string, items []VideoInfo, total int) {
	cacheValue, err := json.Marshal(cachedSearchResult{
		Items:        items,
		TotalResults: total,
	})
	if err != nil {
		return
	}

	s.searchCache.Set(searchKey, string(cacheValue))
}

func buildCacheKey(parts ...string) string {
	return strings.Join(parts, "#")
}
