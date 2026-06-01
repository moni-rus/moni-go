# MoniRus Go SDK

Minimal Sentry-style client for posting events to `POST /api/v1/events`.

## Installation

```bash
go get github.com/moni-rus/moni-go@v0.1.0
```

## Usage

```go
package main

import (
	"context"
	"log"

	monirus "github.com/moni-rus/moni-go"
)

func main() {
	client, err := monirus.NewClientFromDSN("http://localhost:8080/api/v1/events?key=public-key")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.CaptureMessage(
		context.Background(),
		"payment failed",
		monirus.WithTag("component", "checkout"),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```
