package youtube

import (
	"strings"
	"sync"

	"music-bot-v2/internal/application/transport"
)

const (
	apiV3BaseURL = "https://www.googleapis.com/youtube/v3"

	videoEndpoint = "/videos"
)

type Client struct {
	httpClient    *transport.Client
	videoAPIKey   string
	searchKeys    []string
	searchKeyMu   sync.Mutex
	searchKeyNext int
	baseURL       string
}

func NewClient(apiKeys []string, httpClient *transport.Client) *Client {
	if httpClient == nil {
		httpClient = transport.New()
	}

	trimmedKeys := make([]string, 0, len(apiKeys))
	for _, key := range apiKeys {
		if strings.TrimSpace(key) != "" {
			trimmedKeys = append(trimmedKeys, key)
		}
	}

	var videoKey string
	var searchKeys []string
	if len(trimmedKeys) == 1 {
		videoKey = trimmedKeys[0]
	} else if len(trimmedKeys) >= 2 {
		videoKey = trimmedKeys[0]
		searchKeys = append(searchKeys, trimmedKeys[1:]...)
	}

	return &Client{
		httpClient:  httpClient,
		videoAPIKey: videoKey,
		searchKeys:  searchKeys,
		baseURL:     apiV3BaseURL,
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

func (c *Client) nextSearchKey() string {
	c.searchKeyMu.Lock()
	defer c.searchKeyMu.Unlock()
	if len(c.searchKeys) == 0 {
		return c.videoAPIKey
	}
	key := c.searchKeys[c.searchKeyNext]
	c.searchKeyNext++
	if c.searchKeyNext >= len(c.searchKeys) {
		c.searchKeyNext = 0
	}
	return key
}

func (c *Client) videosKey() string {
	if c.videoAPIKey != "" {
		return c.videoAPIKey
	}
	return c.nextSearchKey()
}
