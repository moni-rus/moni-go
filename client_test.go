package monirus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestPackageCaptureMessageFlushesEvent(t *testing.T) {
	received := make(chan Event, 1)
	keys := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload Event
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		keys <- r.Header.Get("X-Monirus-Key")
		received <- payload

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"event_id":"evt-global","queued":true}}`))
	}))
	defer server.Close()

	if err := Init(server.URL+"?key=public-key", WithRelease("api@1.0.0"), WithEnvironment("production")); err != nil {
		t.Fatalf("init: %v", err)
	}

	if err := CaptureMessage("payment failed", WithTag("component", "checkout")); err != nil {
		t.Fatalf("capture message: %v", err)
	}
	if !Flush(time.Second) {
		t.Fatal("flush timed out")
	}

	select {
	case gotKey := <-keys:
		if gotKey != "public-key" {
			t.Fatalf("unexpected key %q", gotKey)
		}
	default:
		t.Fatal("server did not receive api key")
	}

	select {
	case gotPayload := <-received:
		if gotPayload.Message != "payment failed" || gotPayload.Release != "api@1.0.0" || gotPayload.Environment != "production" {
			t.Fatalf("unexpected payload: %#v", gotPayload)
		}
		if gotPayload.Tags["component"] != "checkout" {
			t.Fatalf("missing tag: %#v", gotPayload.Tags)
		}
	default:
		t.Fatal("server did not receive payload")
	}
}

func TestPackageFlushTimesOut(t *testing.T) {
	unblock := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-unblock
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"event_id":"evt-slow","queued":true}}`))
	}))
	defer server.Close()

	if err := Init(server.URL + "?key=public-key"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := CaptureMessage("slow event"); err != nil {
		t.Fatalf("capture message: %v", err)
	}

	if Flush(10 * time.Millisecond) {
		t.Fatal("flush completed before the request was released")
	}

	close(unblock)
	if !Flush(time.Second) {
		t.Fatal("flush timed out after request was released")
	}
}
