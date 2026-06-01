package monirus

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNotInitialized = errors.New("monirus: client is not initialized")

type defaultClientState struct {
	mu      sync.RWMutex
	client  *Client
	flushMu sync.Mutex
	wg      sync.WaitGroup
}

var defaultClient defaultClientState

func Init(dsn string, options ...Option) error {
	client, err := NewClientFromDSN(dsn, options...)
	if err != nil {
		return err
	}

	defaultClient.mu.Lock()
	defaultClient.client = client
	defaultClient.mu.Unlock()

	return nil
}

func CaptureMessage(message string, options ...EventOption) error {
	client := getDefaultClient()
	if client == nil {
		return ErrNotInitialized
	}

	trackDefaultCapture()
	go func() {
		defer defaultClient.wg.Done()
		_, _ = client.CaptureMessage(context.Background(), message, options...)
	}()

	return nil
}

func CaptureException(err error, options ...EventOption) error {
	if err == nil {
		return errors.New("monirus: exception is required")
	}

	client := getDefaultClient()
	if client == nil {
		return ErrNotInitialized
	}

	trackDefaultCapture()
	go func() {
		defer defaultClient.wg.Done()
		_, _ = client.CaptureException(context.Background(), err, options...)
	}()

	return nil
}

func CaptureEvent(event Event, options ...EventOption) error {
	client := getDefaultClient()
	if client == nil {
		return ErrNotInitialized
	}

	trackDefaultCapture()
	go func() {
		defer defaultClient.wg.Done()
		_, _ = client.CaptureEvent(context.Background(), event, options...)
	}()

	return nil
}

func Flush(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		defaultClient.flushMu.Lock()
		defaultClient.wg.Wait()
		defaultClient.flushMu.Unlock()
		close(done)
	}()

	if timeout <= 0 {
		<-done
		return true
	}

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

func getDefaultClient() *Client {
	defaultClient.mu.RLock()
	defer defaultClient.mu.RUnlock()
	return defaultClient.client
}

func trackDefaultCapture() {
	defaultClient.flushMu.Lock()
	defaultClient.wg.Add(1)
	defaultClient.flushMu.Unlock()
}
