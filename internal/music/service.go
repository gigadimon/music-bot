package music

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"

	"music-bot-v2/internal/cacher"
	"music-bot-v2/internal/youtube"
)

type cacherService interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string) error
	DeletePrefix(ctx context.Context, prefix string) error
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

	youtubeClient       youtubeClient
	linkExtractorClient youtubeLinkExtractorClient
}

type VideoInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func NewService(youtubeClient youtubeClient, linkExtractorClient youtubeLinkExtractorClient) *Service {
	return &Service{
		searchCache:         cacher.NewRedis(cacher.SearchCacheDB, 0),
		tokenCache:          cacher.NewRedis(cacher.TokenCacheDB, 0),
		youtubeClient:       youtubeClient,
		linkExtractorClient: linkExtractorClient,
	}
}

func (s *Service) SearchVideos(ctx context.Context, query string, page int, requester string) ([]VideoInfo, int, error) {
	if page < 0 {
		return nil, 0, errors.New("page must be non-negative")
	}

	searchKey := buildCacheKey(requester, query, strconv.Itoa(page))

	if cachedValue, ok, err := s.searchCache.Get(ctx, searchKey); err != nil {
		log.Printf("cache get search key=%s err=%v", searchKey, err)
	} else if ok {
		var cachedResult cachedSearchResult
		if err := json.Unmarshal([]byte(cachedValue), &cachedResult); err == nil {
			return cachedResult.Items, cachedResult.TotalResults, nil
		}
	}

	pageToken := ""
	if page > 0 {
		tokenKey := buildCacheKey(requester, strconv.Itoa(page))
		if cachedToken, ok, err := s.tokenCache.Get(ctx, tokenKey); err != nil {
			log.Printf("cache get token key=%s err=%v", tokenKey, err)
		} else if ok {
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

	go s.storePageTokens(ctx, requester, page, pagination.NextPageToken, pagination.PrevPageToken)
	go s.storeSearch(ctx, searchKey, items, pagination.TotalResults)

	return items, pagination.TotalResults, nil
}

func (s *Service) MP3Link(ctx context.Context, id string) (string, error) {
	return s.linkExtractorClient.MP3Link(ctx, id)
}

func (s *Service) ResetSearchState(ctx context.Context, requester string) {
	if err := s.tokenCache.DeletePrefix(ctx, requester); err != nil {
		log.Printf("cache delete prefix token requester=%s err=%v", requester, err)
	}
	if err := s.searchCache.DeletePrefix(ctx, requester); err != nil {
		log.Printf("cache delete prefix search requester=%s err=%v", requester, err)
	}
}

func (s *Service) storePageTokens(ctx context.Context, requester string, page int, nextToken string, prevToken string) {
	if nextToken != "" {
		if err := s.tokenCache.Set(
			ctx,
			buildCacheKey(requester, strconv.Itoa(page+1)),
			nextToken,
		); err != nil {
			log.Printf("cache set token next requester=%s err=%v", requester, err)
		}
	}

	if prevToken != "" {
		if err := s.tokenCache.Set(
			ctx,
			buildCacheKey(requester, strconv.Itoa(page-1)),
			prevToken,
		); err != nil {
			log.Printf("cache set token prev requester=%s err=%v", requester, err)
		}
	}
}

type cachedSearchResult struct {
	Items        []VideoInfo `json:"items"`
	TotalResults int         `json:"total_results"`
}

func (s *Service) storeSearch(ctx context.Context, searchKey string, items []VideoInfo, total int) {
	cacheValue, err := json.Marshal(cachedSearchResult{
		Items:        items,
		TotalResults: total,
	})
	if err != nil {
		return
	}

	if err := s.searchCache.Set(ctx, searchKey, string(cacheValue)); err != nil {
		log.Printf("cache set search key=%s err=%v", searchKey, err)
	}
}

func buildCacheKey(parts ...string) string {
	return strings.Join(parts, "#")
}
