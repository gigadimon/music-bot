package youtube

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"music-bot-v2/internal/application/transport"
)

const (
	searchEndpoint   = "/search"
	searchMaxResults = 10
)

type Pagination struct {
	NextPageToken string
	PrevPageToken string
	TotalResults  int
}

func (c *Client) Search(ctx context.Context, query string, pageToken string) ([]string, Pagination, error) {
	if strings.TrimSpace(query) == "" {
		return nil, Pagination{}, errors.New("search query is empty")
	}

	params := url.Values{}
	params.Set("key", c.nextSearchKey())
	params.Set("q", query)
	params.Set("type", "video")
	params.Set("maxResults", strconv.Itoa(searchMaxResults))
	if strings.TrimSpace(pageToken) != "" {
		params.Set("pageToken", pageToken)
	}

	request := transport.Request{
		Method: http.MethodGet,
		URL:    c.baseURL + searchEndpoint,
		Query:  params,
	}

	resp, payload, err := transport.DoDecode(ctx, c.httpClient, request, transport.JSONDecoder[searchResponse])
	if err != nil {
		return nil, Pagination{}, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, Pagination{}, fmt.Errorf("search failed: %s", formatAPIError(resp, payload.Error))
	}

	ids := make([]string, 0, len(payload.Items))
	for _, item := range payload.Items {
		if item.ID.Kind == "youtube#video" && item.ID.VideoID != "" {
			ids = append(ids, item.ID.VideoID)
		}
	}

	return ids, Pagination{
		NextPageToken: payload.NextPageToken,
		PrevPageToken: payload.PrevPageToken,
		TotalResults:  payload.PageInfo.TotalResults,
	}, nil
}

type searchResponse struct {
	NextPageToken string `json:"nextPageToken"`
	PrevPageToken string `json:"prevPageToken"`
	Items         []struct {
		ID struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
	PageInfo struct {
		TotalResults int `json:"totalResults"`
	} `json:"pageInfo"`
	Error *apiErrorPayload `json:"error"`
}
