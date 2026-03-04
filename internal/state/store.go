package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type Store struct {
	Path  string
	Clock Clock
}

func (s Store) Load() (FileState, error) {
	path := s.pathOrDefault()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			st := New()
			return st, nil
		}
		return FileState{}, err
	}

	var st FileState
	if err := json.Unmarshal(b, &st); err != nil {
		return FileState{}, fmt.Errorf("invalid state json: %w", err)
	}
	st.Normalize()
	if err := st.Validate(); err != nil {
		return FileState{}, err
	}
	StableSortTransitions(st.LastRun.Transitions)
	return st, nil
}

func (s Store) Save(st FileState) error {
	st.Normalize()
	if st.SchemaVersion == 0 {
		st.SchemaVersion = CurrentSchemaVersion
	}
	if st.PolicyVersion == 0 {
		st.PolicyVersion = DefaultPolicyVersion
	}
	if st.Window == 0 {
		st.Window = DefaultWindow
	}
	StableSortTransitions(st.LastRun.Transitions)
	st.UpdatedAt = s.now().UTC().Format(time.RFC3339)
	if err := st.Validate(); err != nil {
		return err
	}

	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')

	path := s.pathOrDefault()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return atomicWrite(path, b, 0o600)
}

func (s Store) pathOrDefault() string {
	if s.Path == "" {
		return ".flake-state.json"
	}
	return s.Path
}

func (s Store) now() time.Time {
	if s.Clock == nil {
		return realClock{}.Now()
	}
	return s.Clock.Now()
}

func atomicWrite(path string, b []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".flake-state-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = os.Remove(tmpName)
	}
	defer cleanup()

	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	return nil
}
