package runmeta

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFromEnvGitHub(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_REPOSITORY", "org/repo")
	t.Setenv("GITHUB_REF_NAME", "main")
	t.Setenv("GITHUB_SHA", "abc")
	t.Setenv("GITHUB_RUN_ID", "123")
	t.Setenv("GITHUB_REF", "refs/pull/42/merge")
	m := FromEnv()
	if m.CIProvider != "github" || m.RunURL == "" || m.PRNumber != "42" {
		t.Fatalf("unexpected runmeta: %+v", m)
	}
}

func TestFromFileAndMerge(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "runmeta.json")
	if err := os.WriteFile(p, []byte(`{"repo":"f/repo","branch":"dev"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	file, err := FromFile(p)
	if err != nil {
		t.Fatalf("from file: %v", err)
	}
	env := RunMeta{Repo: "env/repo", Commit: "sha"}
	merged := Merge(true, env, file)
	if merged.Repo != "f/repo" || merged.Commit != "sha" {
		t.Fatalf("unexpected merge: %+v", merged)
	}
}
