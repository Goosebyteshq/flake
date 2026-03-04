package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/leomorpho/flake/internal/config"
	"github.com/leomorpho/flake/internal/notify"
	"github.com/leomorpho/flake/internal/state"
)

type NotifyOptions struct {
	StatePath  string
	ConfigPath string
}

type NotifyResult struct {
	Sent       int `json:"sent"`
	Suppressed int `json:"suppressed"`
	Attempted  int `json:"attempted"`
}

type NotifierRunner struct {
	Notifier notify.Notifier
}

func (r NotifierRunner) Run(opts NotifyOptions, stdout io.Writer) (NotifyResult, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return NotifyResult{}, err
	}
	store := state.Store{Path: opts.StatePath}
	st, err := store.Load()
	if err != nil {
		return NotifyResult{}, err
	}
	items := notify.Filter(st, cfg)
	items = notify.ApplySuppressionWithRules(&st, items, cfg.Notify.SuppressRepeatsForRuns, cfg.Notify.MinTransitionAgeRuns, cfg.Notify.OscillationWindowRuns)
	active := notify.Unsuppressed(items)
	if len(active) == 0 {
		fmt.Fprintln(stdout, "notify: no transitions to send")
		return NotifyResult{Sent: 0, Suppressed: len(items), Attempted: len(items)}, nil
	}
	chunks := notify.Chunk(active, cfg.Notify.MaxItemsPerMessage)
	sent := 0
	for _, chunk := range chunks {
		msg := notify.BuildMessage(st.LastRun.RunID, chunk)
		n := r.Notifier
		if n == nil {
			n = notify.SlackWebhookNotifier{
				Webhook: cfg.Slack.Webhook,
				Client:  nil,
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Slack.TimeoutSeconds)*time.Second)
		err := n.Send(ctx, msg)
		cancel()
		if err != nil {
			fmt.Fprintf(stdout, "notify: delivery failed (non-fatal): %v\n", err)
			continue
		}
		sent += len(chunk)
		notify.MarkNotified(&st, chunk)
	}
	_ = store.Save(st)
	fmt.Fprintf(stdout, "notify: sent=%d suppressed=%d attempted=%d\n", sent, len(items)-len(active), len(items))
	return NotifyResult{Sent: sent, Suppressed: len(items) - len(active), Attempted: len(items)}, nil
}
