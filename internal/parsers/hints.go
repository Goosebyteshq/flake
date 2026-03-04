package parsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// HintStore tracks which framework worked for a given repository key.
type HintStore interface {
	Get(repoKey string) (Framework, bool, error)
	Put(repoKey string, framework Framework) error
}

type FileHintStore struct {
	Path string
}

type hintFile struct {
	SchemaVersion int         `json:"schema_version"`
	Repos         []hintEntry `json:"repos"`
}

type hintEntry struct {
	RepoKey   string `json:"repo_key"`
	Framework string `json:"framework"`
}

func (s FileHintStore) Get(repoKey string) (Framework, bool, error) {
	h, err := s.load()
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	target := normalizedRepoKey(repoKey)
	for _, e := range h.Repos {
		if e.RepoKey != target || e.Framework == "" {
			continue
		}
		f, err := ParseFramework(e.Framework)
		if err != nil {
			return "", false, nil
		}
		return f, true, nil
	}
	return "", false, nil
}

func (s FileHintStore) Put(repoKey string, framework Framework) error {
	if framework == FrameworkAuto || framework == "" {
		return fmt.Errorf("invalid framework for hint: %q", framework)
	}
	h, err := s.load()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if h.Repos == nil {
		h = hintFile{SchemaVersion: 1, Repos: []hintEntry{}}
	}
	target := normalizedRepoKey(repoKey)
	replaced := false
	for i := range h.Repos {
		if h.Repos[i].RepoKey == target {
			h.Repos[i].Framework = string(framework)
			replaced = true
			break
		}
	}
	if !replaced {
		h.Repos = append(h.Repos, hintEntry{RepoKey: target, Framework: string(framework)})
	}
	return s.save(h)
}

func (s FileHintStore) load() (hintFile, error) {
	path := strings.TrimSpace(s.Path)
	if path == "" {
		path = ".flake-parser-hints.json"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return hintFile{}, err
	}
	var h hintFile
	if err := json.Unmarshal(b, &h); err != nil {
		return hintFile{}, fmt.Errorf("invalid hint file: %w", err)
	}
	if h.Repos == nil {
		h.Repos = []hintEntry{}
	}
	return h, nil
}

func (s FileHintStore) save(h hintFile) error {
	path := strings.TrimSpace(s.Path)
	if path == "" {
		path = ".flake-parser-hints.json"
	}
	if h.SchemaVersion == 0 {
		h.SchemaVersion = 1
	}

	sort.SliceStable(h.Repos, func(i, j int) bool {
		return h.Repos[i].RepoKey < h.Repos[j].RepoKey
	})

	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	return nil
}

func normalizedRepoKey(repoKey string) string {
	key := strings.TrimSpace(repoKey)
	if key == "" {
		wd, err := os.Getwd()
		if err == nil {
			key = wd
		}
	}
	if abs, err := filepath.Abs(key); err == nil {
		key = abs
	}
	return strings.ToLower(filepath.Clean(key))
}
