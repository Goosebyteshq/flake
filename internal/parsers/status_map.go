package parsers

import "strings"

func mapStatusToken(token string) (TestStatus, bool) {
	switch strings.ToUpper(strings.TrimSpace(token)) {
	case "PASS", "PASSED", "OK", "✓", "✔", "XPASS", "SUCCESS":
		return StatusPass, true
	case "FAIL", "FAILED", "FAILURE", "ERROR", "✕", "✘", "X":
		return StatusFail, true
	case "SKIP", "SKIPPED", "IGNORED", "○", "XFAIL":
		return StatusSkip, true
	default:
		return "", false
	}
}

func mapTRXOutcome(outcome string) (TestStatus, bool) {
	switch strings.ToLower(strings.TrimSpace(outcome)) {
	case "passed":
		return StatusPass, true
	case "failed", "error", "timeout", "aborted":
		return StatusFail, true
	case "notexecuted", "inconclusive", "notrunnable", "warning":
		return StatusSkip, true
	default:
		return "", false
	}
}
