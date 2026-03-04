package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Goosebyteshq/flake/internal/app"
	"github.com/Goosebyteshq/flake/internal/domain"
	"github.com/Goosebyteshq/flake/internal/report"
	"github.com/Goosebyteshq/flake/internal/state"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, app.Scanner{}))
}

func run(args []string, stdout, stderr io.Writer, scanner app.Scanner) int {
	if len(args) < 1 {
		usage(stderr)
		return 2
	}

	switch args[0] {
	case "scan":
		return runScan(args[1:], stdout, stderr, scanner)
	case "report":
		return runReport(args[1:], stdout, stderr)
	case "notify":
		return runNotify(args[1:], stdout, stderr)
	case "migrate":
		return runMigrate(args[1:], stdout, stderr)
	case "publish":
		return runPublish(args[1:], stdout, stderr)
	case "compact":
		return runCompact(args[1:], stdout, stderr)
	default:
		usage(stderr)
		return 2
	}
}

func runScan(args []string, stdout, stderr io.Writer, scanner app.Scanner) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(stderr)
	statePath := fs.String("state", ".flake-state.json", "state file path")
	configPath := fs.String("config", "", "config path")
	window := fs.Int("window", 50, "history window size")
	framework := fs.String("framework", "auto", "framework: auto|go|pytest|jest|junitxml|tap|cargo|trx|surefire|gradle|nunitxml|mocha")
	inputPath := fs.String("input", "", "input file path (defaults to stdin)")
	jsonOut := fs.Bool("json", false, "emit machine readable json")
	eventsPath := fs.String("events", "", "events path (planned for M6)")
	runMetaPath := fs.String("run-meta", "", "run metadata file (planned for M6)")
	debug := fs.Bool("debug", false, "print deterministic parser diagnostics")
	failOn := fs.String("fail-on", "", "reserved: broken|flaky|newfail")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	_ = failOn

	if scanner.Clock == nil {
		scanner.Clock = realClock{}
	}
	result, err := scanner.Run(app.ScanOptions{
		StatePath:       *statePath,
		InputPath:       *inputPath,
		Framework:       *framework,
		Window:          *window,
		EventsPath:      *eventsPath,
		ParserHintsPath: ".flake-parser-hints.json",
		RepoKey:         ".",
		RunMetaPath:     *runMetaPath,
		ConfigPath:      *configPath,
		Debug:           *debug,
	}, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *jsonOut {
		b, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(b))
	}
	triggered := failOnTriggeredClasses(*failOn, result.ClassCounts)
	if len(triggered) > 0 {
		writeFailOnDiagnostics(stderr, *failOn, result, triggered)
		return 1
	}
	return 0
}

func runReport(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(stderr)
	statePath := fs.String("state", ".flake-state.json", "state file path")
	configPath := fs.String("config", "", "config path")
	jsonOut := fs.Bool("json", false, "emit machine readable json")
	view := fs.String("view", "default", "report view: default|unstable|recovered|long-failing")
	limit := fs.Int("limit", 0, "maximum rows to display (0 disables)")
	minFailStreak := fs.Int("min-fail-streak", report.DefaultMinFailStreak, "minimum trailing failures for long-failing view")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	_ = configPath

	summary, err := (app.Reporter{}).Run(app.ReportOptions{
		StatePath:     *statePath,
		View:          *view,
		Limit:         *limit,
		MinFailStreak: *minFailStreak,
	}, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *jsonOut {
		b, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(b))
	}
	return 0
}

func runNotify(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("notify", flag.ContinueOnError)
	fs.SetOutput(stderr)
	statePath := fs.String("state", ".flake-state.json", "state file path")
	configPath := fs.String("config", "", "config path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	_, err := (app.NotifierRunner{}).Run(app.NotifyOptions{
		StatePath:  *statePath,
		ConfigPath: *configPath,
	}, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
	}
	// notify must never fail CI
	return 0
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "flake usage:")
	fmt.Fprintln(w, "  flake scan [flags]")
	fmt.Fprintln(w, "  flake report [flags]")
	fmt.Fprintln(w, "  flake notify [flags]")
	fmt.Fprintln(w, "  flake publish [flags]")
	fmt.Fprintln(w, "  flake compact [flags]")
}

func runMigrate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	statePath := fs.String("state", ".flake-state.json", "state file path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	from, to, err := state.MigrateInPlace(*statePath, realClock{})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "migrate: schema %d -> %d\n", from, to)
	return 0
}

func runPublish(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	fs.SetOutput(stderr)
	eventsPath := fs.String("events", "", "events payload path (or - for stdin)")
	url := fs.String("url", "", "remote publish endpoint URL")
	token := fs.String("token", "", "bearer token")
	tokenEnv := fs.String("token-env", "FLAKE_PUBLISH_TOKEN", "env var containing bearer token")
	timeout := fs.Int("timeout-seconds", 10, "request timeout in seconds")
	retries := fs.Int("retries", 1, "max retry attempts after initial try")
	jsonOut := fs.Bool("json", false, "emit machine readable json")
	debug := fs.Bool("debug", false, "print deterministic publish diagnostics")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	authToken := *token
	if authToken == "" && *tokenEnv != "" {
		authToken = os.Getenv(*tokenEnv)
	}
	result, err := (app.PublisherRunner{}).Run(app.PublishOptions{
		EventsPath:     *eventsPath,
		URL:            *url,
		Token:          authToken,
		TimeoutSeconds: *timeout,
		MaxRetries:     *retries,
		Debug:          *debug,
	}, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *jsonOut {
		b, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(b))
	}
	return 0
}

func runCompact(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("compact", flag.ContinueOnError)
	fs.SetOutput(stderr)
	statePath := fs.String("state", ".flake-state.json", "state file path")
	maxTests := fs.Int("max-tests", 0, "maximum number of tests to retain by recency (0 disables)")
	dropUntouchedDays := fs.Int("drop-untouched-days", 0, "drop tests not seen for N days (0 disables)")
	dryRun := fs.Bool("dry-run", false, "show compact effect without writing state")
	jsonOut := fs.Bool("json", false, "emit machine readable json")
	debug := fs.Bool("debug", false, "print deterministic compact diagnostics")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := (app.Compactor{Clock: realClock{}}).Run(app.CompactOptions{
		StatePath:         *statePath,
		MaxTests:          *maxTests,
		DropUntouchedDays: *dropUntouchedDays,
		DryRun:            *dryRun,
		Debug:             *debug,
	}, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *jsonOut {
		b, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintln(stdout, string(b))
	}
	return 0
}

func shouldFailOn(spec string, counts map[string]int) bool {
	return len(failOnTriggeredClasses(spec, counts)) > 0
}

func failOnTriggeredClasses(spec string, counts map[string]int) []string {
	spec = strings.TrimSpace(strings.ToLower(spec))
	if spec == "" {
		return nil
	}
	out := make([]string, 0, 3)
	seen := map[string]bool{}
	tokens := strings.Split(spec, ",")
	for _, tok := range tokens {
		switch strings.TrimSpace(tok) {
		case "broken":
			if counts["Broken"] > 0 && !seen["Broken"] {
				seen["Broken"] = true
				out = append(out, "Broken")
			}
		case "flaky":
			if counts["Flaky"] > 0 && !seen["Flaky"] {
				seen["Flaky"] = true
				out = append(out, "Flaky")
			}
		case "newfail":
			if counts["NewFail"] > 0 && !seen["NewFail"] {
				seen["NewFail"] = true
				out = append(out, "NewFail")
			}
		}
	}
	return out
}

func writeFailOnDiagnostics(stderr io.Writer, spec string, result app.ScanResult, triggered []string) {
	fmt.Fprintf(stderr, "fail-on triggered: spec=%q classes=%s\n", spec, strings.Join(triggered, ","))

	classSet := map[string]bool{}
	for _, c := range triggered {
		classSet[c] = true
	}
	rows := make([]app.ClassificationDetail, 0, len(result.Classifications))
	for _, row := range result.Classifications {
		if classSet[string(row.Derived.Class)] {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		si := domain.ClassSeverity(rows[i].Derived.Class)
		sj := domain.ClassSeverity(rows[j].Derived.Class)
		if si != sj {
			return si > sj
		}
		if rows[i].Derived.FailureRate != rows[j].Derived.FailureRate {
			return rows[i].Derived.FailureRate > rows[j].Derived.FailureRate
		}
		return rows[i].TestID < rows[j].TestID
	})
	const maxRows = 10
	if len(rows) > maxRows {
		rows = rows[:maxRows]
	}
	for _, row := range rows {
		fmt.Fprintf(stderr, "  %s %0.2f %s reasons=%s\n",
			row.Derived.Class,
			row.Derived.FailureRate,
			row.TestID,
			strings.Join(row.Explanation.Reasons, ","),
		)
	}
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now()
}
