package publish

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPPublisherRetryThenSuccess(t *testing.T) {
	var calls int32
	var firstIDKey string
	var secondIDKey string
	var firstRunID string
	var secondRunID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&calls, 1)
		if c == 1 {
			firstIDKey = r.Header.Get("X-Flake-Idempotency-Key")
			firstRunID = r.Header.Get("X-Flake-Run-ID")
		} else {
			secondIDKey = r.Header.Get("X-Flake-Idempotency-Key")
			secondRunID = r.Header.Get("X-Flake-Run-ID")
		}
		if c == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := HTTPPublisher{
		URL:        srv.URL,
		Token:      "x",
		Client:     srv.Client(),
		MaxRetries: 1,
		RetryDelay: 1 * time.Millisecond,
	}
	payload := []byte(`{"run":{"run_id":"run-123"},"ok":true}`)
	if err := p.Publish(context.Background(), payload); err != nil {
		t.Fatalf("Publish error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("calls=%d, want 2", calls)
	}
	expected := sha256.Sum256(payload)
	expectedHex := hex.EncodeToString(expected[:])
	if firstIDKey == "" || secondIDKey == "" {
		t.Fatalf("missing idempotency key header")
	}
	if firstIDKey != secondIDKey || firstIDKey != expectedHex {
		t.Fatalf("idempotency key mismatch first=%q second=%q expected=%q", firstIDKey, secondIDKey, expectedHex)
	}
	if firstRunID != "run-123" || secondRunID != "run-123" {
		t.Fatalf("run id header mismatch first=%q second=%q", firstRunID, secondRunID)
	}
}

func TestHTTPPublisherFailsAfterRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := HTTPPublisher{URL: srv.URL, Client: srv.Client(), MaxRetries: 1, RetryDelay: 1 * time.Millisecond}
	if err := p.Publish(context.Background(), []byte(`{}`)); err == nil {
		t.Fatalf("expected publish error")
	}
}

func TestHTTPPublisherOmitsRunHeaderWhenMissing(t *testing.T) {
	var runHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runHeader = r.Header.Get("X-Flake-Run-ID")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	p := HTTPPublisher{URL: srv.URL, Client: srv.Client()}
	if err := p.Publish(context.Background(), []byte(`{"schema_version":1}`)); err != nil {
		t.Fatalf("Publish error: %v", err)
	}
	if strings.TrimSpace(runHeader) != "" {
		t.Fatalf("expected empty run header, got %q", runHeader)
	}
}

func TestHTTPPublisherAttemptHookDeterministic(t *testing.T) {
	var calls int32
	attempts := make([]string, 0)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&calls, 1)
		if c == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("upstream bad gateway"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := HTTPPublisher{
		URL:        srv.URL,
		Client:     srv.Client(),
		MaxRetries: 1,
		RetryDelay: 1 * time.Millisecond,
		OnAttempt: func(ev AttemptEvent) {
			attempts = append(attempts, fmt.Sprintf("%d/%d:%d:%t", ev.Attempt, ev.Max, ev.Status, ev.WillRetry))
		},
	}
	if err := p.Publish(context.Background(), []byte(`{"run":{"run_id":"r"}}`)); err != nil {
		t.Fatalf("publish error: %v", err)
	}
	if len(attempts) != 2 {
		t.Fatalf("attempts len=%d want=2", len(attempts))
	}
	if attempts[0] != "1/2:502:true" || attempts[1] != "2/2:200:false" {
		t.Fatalf("unexpected attempt log: %+v", attempts)
	}
}
