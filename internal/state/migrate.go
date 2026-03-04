package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type migrationProbe struct {
	SchemaVersion int `json:"schema_version"`
}

// MigrateInPlace upgrades a state file to the current schema version.
// Returns fromVersion and toVersion.
func MigrateInPlace(path string, clock Clock) (int, int, error) {
	store := Store{Path: path, Clock: clock}
	b, err := os.ReadFile(store.pathOrDefault())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			st := New()
			if err := store.Save(st); err != nil {
				return 0, 0, err
			}
			return 0, CurrentSchemaVersion, nil
		}
		return 0, 0, err
	}

	var probe migrationProbe
	_ = json.Unmarshal(b, &probe)
	from := probe.SchemaVersion

	switch from {
	case 0:
		var legacy legacyV0
		if err := json.Unmarshal(b, &legacy); err != nil {
			return 0, 0, fmt.Errorf("invalid legacy state json: %w", err)
		}
		migrated := FileState{
			SchemaVersion: CurrentSchemaVersion,
			PolicyVersion: legacy.PolicyVersion,
			Window:        legacy.Window,
			UpdatedAt:     legacy.UpdatedAt,
			Tests:         legacy.Tests,
			LastRun:       legacy.LastRun,
			LastRunIndex:  0,
		}
		migrated.Normalize()
		if migrated.PolicyVersion == 0 {
			migrated.PolicyVersion = DefaultPolicyVersion
		}
		if migrated.Window == 0 {
			migrated.Window = DefaultWindow
		}
		if err := store.Save(migrated); err != nil {
			return 0, 0, err
		}
		return 0, CurrentSchemaVersion, nil
	case CurrentSchemaVersion:
		st, err := store.Load()
		if err != nil {
			return from, from, err
		}
		if err := store.Save(st); err != nil {
			return from, from, err
		}
		return from, from, nil
	default:
		return from, 0, fmt.Errorf("unsupported schema_version=%d", from)
	}
}

type legacyV0 struct {
	PolicyVersion int                 `json:"policy_version"`
	Window        int                 `json:"window"`
	UpdatedAt     string              `json:"updated_at"`
	Tests         map[string]TestSlot `json:"tests"`
	LastRun       LastRun             `json:"last_run"`
}
