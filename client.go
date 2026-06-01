package monirus

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const defaultEndpoint = "http://localhost:8080/api/v1/events"

type Client struct {
	endpoint    string
	apiKey      string
	httpClient  *http.Client
	release     string
	environment string
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

func WithRelease(release string) Option {
	return func(client *Client) {
		client.release = release
	}
}

func WithEnvironment(environment string) Option {
	return func(client *Client) {
		client.environment = environment
	}
}

func NewClient(endpoint string, apiKey string, options ...Option) (*Client, error) {
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	if apiKey == "" {
		return nil, errors.New("monirus: api key is required")
	}

	client := &Client{
		endpoint:   endpoint,
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}

	for _, option := range options {
		option(client)
	}

	return client, nil
}

type Event struct {
	EventID     string           `json:"event_id,omitempty"`
	Type        string           `json:"type,omitempty"`
	Level       string           `json:"level,omitempty"`
	Message     string           `json:"message,omitempty"`
	Environment string           `json:"environment,omitempty"`
	Release     string           `json:"release,omitempty"`
	Timestamp   *time.Time       `json:"timestamp,omitempty"`
	Fingerprint []string         `json:"fingerprint,omitempty"`
	Exception   *Exception       `json:"exception,omitempty"`
	User        map[string]any   `json:"user,omitempty"`
	Tags        map[string]any   `json:"tags,omitempty"`
	Request     map[string]any   `json:"request,omitempty"`
	Breadcrumbs []map[string]any `json:"breadcrumbs,omitempty"`
}

type Exception struct {
	Type       string         `json:"type,omitempty"`
	Value      string         `json:"value,omitempty"`
	Stacktrace map[string]any `json:"stacktrace,omitempty"`
}

type Response struct {
	EventID string `json:"event_id"`
	Queued  bool   `json:"queued"`
}

type apiResponse struct {
	Data Response `json:"data"`
}

func (client *Client) CaptureMessage(ctx context.Context, message string, options ...EventOption) (Response, error) {
	event := Event{
		Type:    "error",
		Level:   "error",
		Message: message,
	}

	return client.CaptureEvent(ctx, event, options...)
}

func (client *Client) CaptureException(ctx context.Context, err error, options ...EventOption) (Response, error) {
	if err == nil {
		return Response{}, errors.New("monirus: exception is required")
	}

	event := Event{
		Type:    "error",
		Level:   "error",
		Message: err.Error(),
		Exception: &Exception{
			Type:  fmt.Sprintf("%T", err),
			Value: err.Error(),
		},
	}

	return client.CaptureEvent(ctx, event, options...)
}

func (client *Client) CaptureEvent(ctx context.Context, event Event, options ...EventOption) (Response, error) {
	for _, option := range options {
		option(&event)
	}

	if event.Environment == "" {
		event.Environment = client.environment
	}
	if event.Release == "" {
		event.Release = client.release
	}

	body, err := json.Marshal(event)
	if err != nil {
		return Response{}, fmt.Errorf("monirus: encode event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("monirus: build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Monirus-Key", client.apiKey)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("monirus: send event: %w", err)
	}
	defer resp.Body.Close()

	var payload apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Response{}, fmt.Errorf("monirus: decode response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("monirus: server returned %s", resp.Status)
	}

	if payload.Data.EventID == "" || !payload.Data.Queued {
		return Response{}, errors.New("monirus: response did not contain a queued event")
	}

	return payload.Data, nil
}

func (client *Client) Endpoint() string {
	return client.endpoint
}

func (client *Client) APIKey() string {
	return client.apiKey
}

func endpointFromDSN(dsn string) (string, string) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return dsn, ""
	}

	key := parsed.Query().Get("key")
	parsed.RawQuery = ""

	return parsed.String(), key
}
