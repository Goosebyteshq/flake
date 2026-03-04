package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Goosebyteshq/flake/internal/events"
	"github.com/Goosebyteshq/flake/internal/publish"
)

type PublishOptions struct {
	EventsPath     string
	URL            string
	Token          string
	TimeoutSeconds int
	MaxRetries     int
	Debug          bool
}

type PublishResult struct {
	Published     bool     `json:"published"`
	Bytes         int      `json:"bytes"`
	DebugMessages []string `json:"debug_messages,omitempty"`
}

type PublisherRunner struct {
	Publisher interface {
		Publish(ctx context.Context, payload []byte) error
	}
}

func (r PublisherRunner) Run(opts PublishOptions, stdout io.Writer) (PublishResult, error) {
	payload, err := readPayload(opts.EventsPath)
	if err != nil {
		return PublishResult{}, err
	}
	// Validate contract before sending.
	var p events.Payload
	if err := json.Unmarshal(payload, &p); err != nil {
		return PublishResult{}, fmt.Errorf("invalid events payload: %w", err)
	}
	if p.SchemaVersion <= 0 {
		return PublishResult{}, fmt.Errorf("invalid events payload: schema_version is required")
	}
	debug := make([]string, 0)

	timeout := opts.TimeoutSeconds
	if timeout <= 0 {
		timeout = 10
	}
	pub := r.Publisher
	if pub == nil {
		httpPub := publish.HTTPPublisher{
			URL:        opts.URL,
			Token:      opts.Token,
			MaxRetries: opts.MaxRetries,
		}
		if opts.Debug {
			httpPub.OnAttempt = func(ev publish.AttemptEvent) {
				msg := fmt.Sprintf("attempt=%d/%d status=%d retry=%t", ev.Attempt, ev.Max, ev.Status, ev.WillRetry)
				if ev.Err != "" {
					msg += " err=" + sanitizeErr(ev.Err)
				}
				debug = append(debug, msg)
			}
		}
		pub = httpPub
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	if err := pub.Publish(ctx, payload); err != nil {
		return PublishResult{DebugMessages: debug}, err
	}
	if opts.Debug {
		debug = append(debug, fmt.Sprintf("endpoint=%s payload_bytes=%d", sanitizeEndpoint(opts.URL), len(payload)))
		fmt.Fprintf(stdout, "debug: %s\n", strings.Join(debug, " | "))
	}
	fmt.Fprintf(stdout, "publish: success bytes=%d\n", len(payload))
	return PublishResult{Published: true, Bytes: len(payload), DebugMessages: debug}, nil
}

func readPayload(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("--events is required")
	}
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func sanitizeEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if i := strings.Index(raw, "?"); i >= 0 {
		raw = raw[:i]
	}
	return raw
}

func sanitizeErr(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return ""
	}
	// Keep concise deterministic error signal in debug output.
	if len(msg) > 120 {
		return msg[:120]
	}
	return msg
}
