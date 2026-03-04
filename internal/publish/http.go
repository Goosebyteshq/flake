package publish

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPPublisher struct {
	URL        string
	Token      string
	Client     *http.Client
	MaxRetries int
	RetryDelay time.Duration
	OnAttempt  func(AttemptEvent)
}

type AttemptEvent struct {
	Attempt   int
	Max       int
	Status    int
	Err       string
	WillRetry bool
}

func (p HTTPPublisher) Publish(ctx context.Context, payload []byte) error {
	url := strings.TrimSpace(p.URL)
	if url == "" {
		return fmt.Errorf("publish url is required")
	}
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	retries := p.MaxRetries
	if retries < 0 {
		retries = 0
	}
	delay := p.RetryDelay
	if delay <= 0 {
		delay = 200 * time.Millisecond
	}

	attempts := retries + 1
	idempotencyKey := payloadHash(payload)
	runID := extractRunID(payload)
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if p.Token != "" {
			req.Header.Set("Authorization", "Bearer "+p.Token)
		}
		req.Header.Set("X-Flake-Idempotency-Key", idempotencyKey)
		if runID != "" {
			req.Header.Set("X-Flake-Run-ID", runID)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if p.OnAttempt != nil {
				p.OnAttempt(AttemptEvent{
					Attempt:   attempt,
					Max:       attempts,
					Status:    0,
					Err:       err.Error(),
					WillRetry: attempt < attempts,
				})
			}
		} else {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				if p.OnAttempt != nil {
					p.OnAttempt(AttemptEvent{
						Attempt:   attempt,
						Max:       attempts,
						Status:    resp.StatusCode,
						Err:       "",
						WillRetry: false,
					})
				}
				return nil
			}
			lastErr = fmt.Errorf("publish http status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
			if p.OnAttempt != nil {
				p.OnAttempt(AttemptEvent{
					Attempt:   attempt,
					Max:       attempts,
					Status:    resp.StatusCode,
					Err:       strings.TrimSpace(string(body)),
					WillRetry: attempt < attempts,
				})
			}
		}

		if attempt < attempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return lastErr
}

func payloadHash(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func extractRunID(payload []byte) string {
	var probe struct {
		Run struct {
			RunID string `json:"run_id"`
		} `json:"run"`
	}
	if err := json.Unmarshal(payload, &probe); err != nil {
		return ""
	}
	return strings.TrimSpace(probe.Run.RunID)
}
