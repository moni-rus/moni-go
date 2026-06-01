package monirus

import "errors"

func NewClientFromDSN(dsn string, options ...Option) (*Client, error) {
	endpoint, apiKey := endpointFromDSN(dsn)
	if apiKey == "" {
		return nil, errors.New("monirus: DSN must include ?key=")
	}

	return NewClient(endpoint, apiKey, options...)
}
