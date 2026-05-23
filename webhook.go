package haybtech

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	MaxPayloadSize = 1024 * 1024 // 1MB
	DefaultTolerance = 300 // 5 minutes
)

func ConstructEvent(payload []byte, sigHeader string, secret string) (map[string]interface{}, error) {
	if len(payload) > MaxPayloadSize {
		return nil, fmt.Errorf("payload too large")
	}

	parts := make(map[string]string)
	for _, part := range strings.Split(sigHeader, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			parts[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	tStr, okT := parts["t"]
	v1, okV := parts["v1"]
	if !okT || !okV {
		return nil, fmt.Errorf("malformed signature header")
	}

	// Parse timestamp
	var timestamp int64
	fmt.Sscanf(tStr, "%d", &timestamp)

	// Replay protection
	if math.Abs(float64(time.Now().Unix()-timestamp)) > DefaultTolerance {
		return nil, fmt.Errorf("webhook signature expired")
	}

	// Verify signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.%s", timestamp, string(payload))))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison
	if !hmac.Equal([]byte(expectedSig), []byte(v1)) {
		return nil, fmt.Errorf("invalid signature")
	}

	var event map[string]interface{}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}

	return event, nil
}
