# MoniRus Go SDK

Minimal Sentry-style client for posting events to `POST /api/v1/events`.

## Installation

```bash
go get github.com/moni-rus/moni-go
```

## Usage

```bash
export MONIRUS_DSN="http://localhost:8080/api/v1/events?key=public-key"
```

```go
package main

import (
	"context"
	"log"
	"os"

	monirus "github.com/moni-rus/moni-go"
)

func main() {
	client, err := monirus.NewClientFromDSN(os.Getenv("MONIRUS_DSN"))
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
