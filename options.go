package monirus

import "time"

type EventOption func(*Event)

func WithEventID(eventID string) EventOption {
	return func(event *Event) {
		event.EventID = eventID
	}
}

func WithLevel(level string) EventOption {
	return func(event *Event) {
		event.Level = level
	}
}

func WithTag(key string, value any) EventOption {
	return func(event *Event) {
		if event.Tags == nil {
			event.Tags = map[string]any{}
		}
		event.Tags[key] = value
	}
}

func WithUser(user map[string]any) EventOption {
	return func(event *Event) {
		event.User = user
	}
}

func WithTimestamp(timestamp time.Time) EventOption {
	return func(event *Event) {
		event.Timestamp = &timestamp
	}
}

func WithFingerprint(fingerprint ...string) EventOption {
	return func(event *Event) {
		event.Fingerprint = fingerprint
	}
}
