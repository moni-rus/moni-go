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
	"log"
	"os"
	"time"

	monirus "github.com/moni-rus/moni-go"
)

func main() {
	if err := monirus.Init(os.Getenv("MONIRUS_DSN")); err != nil {
		log.Fatal(err)
	}
	defer monirus.Flush(2 * time.Second)

	if err := monirus.CaptureMessage("payment failed", monirus.WithTag("component", "checkout")); err != nil {
		log.Fatal(err)
	}
}
```
