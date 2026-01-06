package yt1s

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"music-bot-v2/internal/application/transport"
)

const (
	endpoint = "https://ac.insvid.com/converter"
	origin   = "https://ac.insvid.com"
)

type Client struct {
	httpClient *transport.Client
}

func NewClient(httpClient *transport.Client) *Client {
	if httpClient == nil {
		httpClient = transport.New()
	}

	return &Client{httpClient: httpClient}
}

type requestPayload struct {
	ID       string `json:"id"`
	FileType string `json:"fileType"`
}

type responsePayload struct {
	Link string `json:"link"`
}

func (c *Client) MP3Link(ctx context.Context, id string) (string, error) {
	if strings.TrimSpace(id) == "" {
		return "", errors.New("id is empty")
	}

	request := transport.Request{
		Method: http.MethodPost,
		URL:    endpoint,
		Body: requestPayload{
			ID:       id,
			FileType: "MP3",
		},
		Encoder: transport.JSONEncoder,
		Headers: map[string][]string{
			"Origin": {origin},
		},
	}

	resp, payload, err := transport.DoDecode(ctx, c.httpClient, request, transport.JSONDecoder[responsePayload])
	if err != nil {
		return "", err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("request failed: %s", formatAPIError(resp))
	}

	link := strings.TrimSpace(payload.Link)
	if link == "" {
		return "", errors.New("link is empty")
	}

	return link, nil
}

func formatAPIError(resp transport.Response) string {
	if len(resp.Body) > 0 {
		body := strings.TrimSpace(string(resp.Body))
		if body != "" {
			return body
		}
	}
	return resp.Status
}
