package youtube

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"music-bot-v2/internal/application/transport"
)

var isoDurationRE = regexp.MustCompile(`^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$`)

func (c *Client) Videos(ctx context.Context, ids []string) (map[string]string, error) {
	if len(ids) == 0 {
		return nil, errors.New("video ids are empty")
	}

	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("id", strings.Join(ids, ","))
	params.Set("part", "snippet,contentDetails")

	request := transport.Request{
		Method: http.MethodGet,
		URL:    c.baseURL + videoEndpoint,
		Query:  params,
	}

	resp, payload, err := transport.DoDecode(ctx, c.httpClient, request, transport.JSONDecoder[videosResponse])
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("videos failed: %s", formatAPIError(resp, payload.Error))
	}

	results := make(map[string]string, len(payload.Items))
	for _, item := range payload.Items {
		if item.Kind != "youtube#video" {
			continue
		}
		formatted, err := formatDuration(item.ContentDetails.Duration)
		if err != nil {
			return nil, err
		}
		results[item.ID] = fmt.Sprintf("%s %s", formatted, item.Snippet.Title)
	}

	return results, nil
}

type videosResponse struct {
	Items []struct {
		Kind    string `json:"kind"`
		ID      string `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
		ContentDetails struct {
			Duration string `json:"duration"`
		} `json:"contentDetails"`
	} `json:"items"`
	Error *apiErrorPayload `json:"error"`
}

func formatDuration(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("duration is empty")
	}

	matches := isoDurationRE.FindStringSubmatch(raw)
	if matches == nil {
		return "", fmt.Errorf("invalid duration format: %s", raw)
	}

	hours, err := parseDurationPart(matches[1])
	if err != nil {
		return "", err
	}
	minutes, err := parseDurationPart(matches[2])
	if err != nil {
		return "", err
	}
	seconds, err := parseDurationPart(matches[3])
	if err != nil {
		return "", err
	}

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds), nil
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds), nil
}

func parseDurationPart(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid duration part: %s", raw)
	}
	return value, nil
}
