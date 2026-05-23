# HayBTech Go SDK

Official Go SDK for the HayBTech Payment Gateway API -- mobile payments across West Africa .

[![Go Reference](https://pkg.go.dev/badge/github.com/haybtech/haybtech-go-sdk.svg)](https://pkg.go.dev/github.com/haybtech/haybtech-go-sdk)
[![Go](https://img.shields.io/badge/go-1.18+-00ADD8.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## Installation

```bash
go get github.com/haybtech/haybtech-go-sdk
```

---

## Quick Start

Initialize the client with your secret key (`sk_live_...` or `sk_test_...`):

```go
package main

import (
    "fmt"
    "log"

    haybtech "github.com/haybtech/haybtech-go-sdk"
)

func main() {
    client, err := haybtech.NewClient("sk_test_your_key")
    if err != nil {
        log.Fatal(err)
    }

    // Initiate a payment
    response, err := client.Payments.Create(map[string]interface{}{
        "merchant_ref": "ORDER-12345",
        "amount":       5000,
        "currency":     "XOF",
        "return_url":   "https://mysite.com/success",
        "cancel_url":   "https://mysite.com/cancel",
        "callback_url": "https://mysite.com/webhook",
    })
    if err != nil {
        log.Fatal(err)
    }

    data := response["data"].(map[string]interface{})
    fmt.Println("Payment URL:", data["payment_url"])
}
```

---

## Webhooks (net/http)

Securely verify incoming webhooks from HayBTech:

```go
package main

import (
    "fmt"
    "io"
    "net/http"

    haybtech "github.com/haybtech/haybtech-go-sdk"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    payload, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    signature := r.Header.Get("X-HayBTech-Signature")
    secret := "whsec_..."

    event, err := haybtech.ConstructEvent(payload, signature, secret)
    if err != nil {
        http.Error(w, "Invalid Signature", http.StatusForbidden)
        return
    }

    switch event["event"] {
    case "payment.success":
        merchantRef := event["data"].(map[string]interface{})["merchant_ref"]
        fmt.Println("Payment confirmed for:", merchantRef)
        // Mark order as paid
    case "payment.failed":
        // Handle failure
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "OK")
}

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    http.ListenAndServe(":8080", nil)
}
```

### With Gin

```go
import "github.com/gin-gonic/gin"

func webhookHandler(c *gin.Context) {
    payload, _ := io.ReadAll(c.Request.Body)
    signature := c.GetHeader("X-HayBTech-Signature")

    event, err := haybtech.ConstructEvent(payload, signature, "whsec_...")
    if err != nil {
        c.JSON(403, gin.H{"error": "Invalid Signature"})
        return
    }

    if event["event"] == "payment.success" {
        // Mark order as paid
    }

    c.JSON(200, gin.H{"status": "ok"})
}
```

---

## Available Events

| Event                     | Description              |
|:--------------------------|:-------------------------|
| `payment.success`         | Payment confirmed        |
| `payment.failed`          | Payment failed           |
| `payment.cancelled`       | Cancelled by customer    |
| `payment.expired`         | Payment timed out        |

---

## Error Handling

```go
response, err := client.Payments.Create(params)
if err != nil {
    // Errors contain the HTTP status code and API response body
    // e.g., "API error (422): {"error":"insufficient_funds"}"
    log.Println("Payment failed:", err)
    return
}
```

---

## Test Mode

```go
client, _ := haybtech.NewClient("sk_test_...") // No real charges
```

---

## Advanced Usage

```go
// Custom timeout (the test/live mode is determined by your key, not the URL)
client, _ := haybtech.NewClient("sk_test_...")
client.HTTPClient.Timeout = 60 * time.Second
```

---

## Security Features

This SDK is built for **Maximum Security**:

- **Zero Dependencies**: Uses only the Go standard library (`net/http`, `crypto/hmac`, `crypto/sha256`). No third-party modules to compromise via supply chain attacks.
- **Secret Masking**: Keys are automatically masked in `String()` output to prevent accidental log exposure.
- **Memory Protection**: Webhook payloads are capped at 1 MB to prevent memory exhaustion (DoS).
- **Timing Attack Resistance**: Uses `hmac.Equal()` for constant-time HMAC signature comparison.
- **Replay Protection**: Webhook timestamps are validated within a 5-minute tolerance window.
- **CRLF Guard**: Prevents HTTP header injection via malformed API keys.

---

## Requirements

| Requirement | Version |
|:------------|:--------|
| Go          | 1.18+   |

---

## API Resources

- `client.Payments` -- Create, retrieve, list, and verify transactions.

---

MIT License
