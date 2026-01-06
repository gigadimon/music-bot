package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	headers    http.Header
}

type Option func(*Client)

func New(options ...Option) *Client {
	client := &Client{
		httpClient: &http.Client{},
		headers:    make(http.Header),
	}
	for _, option := range options {
		option(client)
	}
	return client
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		client.httpClient.Timeout = timeout
	}
}

func WithHeader(key, value string) Option {
	return func(client *Client) {
		client.headers.Set(key, value)
	}
}

func WithHeaders(headers http.Header) Option {
	return func(client *Client) {
		for key, values := range headers {
			client.headers.Del(key)
			for _, value := range values {
				client.headers.Add(key, value)
			}
		}
	}
}

type Request struct {
	Method      string
	URL         string
	Headers     http.Header
	Query       url.Values
	Body        any
	ContentType string
	Encoder     Encoder
}

type Response struct {
	Status     string
	StatusCode int
	Headers    http.Header
	Body       []byte
}

type (
	Encoder        func(ctx context.Context, body any) (io.ReadCloser, string, error)
	Decoder[T any] func(ctx context.Context, resp Response) (T, error)
)

func (c *Client) Do(ctx context.Context, request Request) (Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	method := request.Method
	if method == "" {
		method = http.MethodGet
	}

	parsedURL, err := url.Parse(request.URL)
	if err != nil {
		return Response{}, err
	}

	if len(request.Query) > 0 {
		query := parsedURL.Query()
		for key, values := range request.Query {
			for _, value := range values {
				query.Add(key, value)
			}
		}
		parsedURL.RawQuery = query.Encode()
	}

	body, contentType, err := c.encodeBody(ctx, request)
	if err != nil {
		return Response{}, err
	}
	if body != nil {
		defer body.Close()
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), body)
	if err != nil {
		return Response{}, err
	}

	headers := cloneHeaders(c.headers)
	mergeHeaders(headers, request.Headers)

	if request.ContentType != "" {
		headers.Set("Content-Type", request.ContentType)
	} else if contentType != "" && headers.Get("Content-Type") == "" {
		headers.Set("Content-Type", contentType)
	}
	httpRequest.Header = headers

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Response{}, err
	}

	return Response{
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Headers:    response.Header.Clone(),
		Body:       responseBody,
	}, nil
}

func DoDecode[T any](ctx context.Context, client *Client, request Request, decoder Decoder[T]) (Response, T, error) {
	if decoder == nil {
		var zero T
		return Response{}, zero, errors.New("decoder is nil")
	}

	resp, err := client.Do(ctx, request)
	if err != nil {
		var zero T
		return Response{}, zero, err
	}

	decoded, err := decoder(ctx, resp)
	if err != nil {
		var zero T
		return resp, zero, err
	}

	return resp, decoded, nil
}

func Decode[T any](ctx context.Context, resp Response, decoder Decoder[T]) (T, error) {
	if decoder == nil {
		var zero T
		return zero, errors.New("decoder is nil")
	}

	return decoder(ctx, resp)
}

func JSONEncoder(ctx context.Context, body any) (io.ReadCloser, string, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}

	return io.NopCloser(bytes.NewReader(payload)), "application/json", nil
}

func JSONDecoder[T any](ctx context.Context, resp Response) (T, error) {
	var out T
	if len(resp.Body) == 0 {
		return out, nil
	}

	if err := json.Unmarshal(resp.Body, &out); err != nil {
		return out, err
	}

	return out, nil
}

func (c *Client) encodeBody(ctx context.Context, request Request) (io.ReadCloser, string, error) {
	if request.Body == nil {
		return nil, "", nil
	}

	if request.Encoder != nil {
		return request.Encoder(ctx, request.Body)
	}

	switch body := request.Body.(type) {
	case io.ReadCloser:
		return body, "", nil
	case io.Reader:
		return io.NopCloser(body), "", nil
	case []byte:
		return io.NopCloser(bytes.NewReader(body)), "", nil
	case string:
		return io.NopCloser(strings.NewReader(body)), "", nil
	case url.Values:
		encoded := body.Encode()
		return io.NopCloser(strings.NewReader(encoded)), "application/x-www-form-urlencoded", nil
	default:
		return nil, "", errors.New("unsupported body type without encoder")
	}
}

func mergeHeaders(target http.Header, extra http.Header) {
	for key, values := range extra {
		target.Del(key)
		for _, value := range values {
			target.Add(key, value)
		}
	}
}

func cloneHeaders(headers http.Header) http.Header {
	if headers == nil {
		return make(http.Header)
	}
	return headers.Clone()
}
