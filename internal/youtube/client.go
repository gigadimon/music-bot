package youtube

import (
	"strings"

	"music-bot-v2/internal/application/transport"
)

const (
	apiV3BaseURL = "https://www.googleapis.com/youtube/v3"

	videoEndpoint = "/videos"
)

type Client struct {
	httpClient *transport.Client
	apiKey     string
	baseURL    string
}

func NewClient(apiKey string, httpClient *transport.Client) *Client {
	if httpClient == nil {
		httpClient = transport.New()
	}

	return &Client{
		httpClient: httpClient,
		apiKey:     apiKey,
		baseURL:    apiV3BaseURL,
	}
}

type apiErrorPayload struct {
	Message string `json:"message"`
}

func formatAPIError(resp transport.Response, payload *apiErrorPayload) string {
	if payload != nil && strings.TrimSpace(payload.Message) != "" {
		return payload.Message
	}
	if len(resp.Body) > 0 {
		body := strings.TrimSpace(string(resp.Body))
		if body != "" {
			return body
		}
	}
	return resp.Status
}
