package monirus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCaptureMessageQueuesEvent(t *testing.T) {
	var gotKey string
	var gotPayload Event

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Monirus-Key")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"event_id":"evt-go","queued":true}}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "public-key", WithRelease("api@1.0.0"), WithEnvironment("production"))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	response, err := client.CaptureMessage(context.Background(), "payment failed", WithTag("component", "checkout"))
	if err != nil {
		t.Fatalf("capture message: %v", err)
	}

	if response.EventID != "evt-go" || !response.Queued {
		t.Fatalf("unexpected response: %#v", response)
	}
	if gotKey != "public-key" {
		t.Fatalf("unexpected key %q", gotKey)
	}
	if gotPayload.Message != "payment failed" || gotPayload.Release != "api@1.0.0" || gotPayload.Environment != "production" {
		t.Fatalf("unexpected payload: %#v", gotPayload)
	}
	if gotPayload.Tags["component"] != "checkout" {
		t.Fatalf("missing tag: %#v", gotPayload.Tags)
	}
}

func TestNewClientFromDSN(t *testing.T) {
	client, err := NewClientFromDSN("http://monirus.test/api/v1/events?key=abc123")
	if err != nil {
		t.Fatalf("new client from dsn: %v", err)
	}

	if client.Endpoint() != "http://monirus.test/api/v1/events" {
		t.Fatalf("unexpected endpoint %q", client.Endpoint())
	}
	if client.APIKey() != "abc123" {
		t.Fatalf("unexpected api key %q", client.APIKey())
	}
}
