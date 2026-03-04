package parsers

import "io"

// Parser reads test output and emits normalized results.
type Parser interface {
	Name() Framework
	Detect(sample []byte) bool
	Parse(r io.Reader) ([]TestResult, error)
}

// ConfidenceParser can provide a confidence score in [0,100].
// Higher scores are tried first during auto-detection.
type ConfidenceParser interface {
	DetectConfidence(sample []byte) int
}

type CandidateInfo struct {
	Framework Framework `json:"framework"`
	Score     int       `json:"score"`
	Source    string    `json:"source"` // hint|detect|fallback|explicit
}
