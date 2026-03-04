package runmeta

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type RunMeta struct {
	Repo       string `json:"repo"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	CIProvider string `json:"ci_provider"`
	CIRunID    string `json:"ci_run_id"`
	RunURL     string `json:"run_url"`
	PRNumber   string `json:"pr_number"`
}

func (m RunMeta) ToMap() map[string]string {
	out := map[string]string{}
	add := func(k, v string) {
		if strings.TrimSpace(v) != "" {
			out[k] = v
		}
	}
	add("repo", m.Repo)
	add("branch", m.Branch)
	add("commit", m.Commit)
	add("ci_provider", m.CIProvider)
	add("ci_run_id", m.CIRunID)
	add("run_url", m.RunURL)
	add("pr_number", m.PRNumber)
	return out
}

func FromEnv() RunMeta {
	if v := strings.TrimSpace(os.Getenv("GITHUB_ACTIONS")); v == "true" {
		repo := os.Getenv("GITHUB_REPOSITORY")
		runID := os.Getenv("GITHUB_RUN_ID")
		return RunMeta{
			Repo:       repo,
			Branch:     os.Getenv("GITHUB_REF_NAME"),
			Commit:     os.Getenv("GITHUB_SHA"),
			CIProvider: "github",
			CIRunID:    runID,
			RunURL:     githubRunURL(repo, runID),
			PRNumber:   githubPRFromRef(os.Getenv("GITHUB_REF")),
		}
	}
	if v := strings.TrimSpace(os.Getenv("GITLAB_CI")); v == "true" {
		return RunMeta{
			Repo:       os.Getenv("CI_PROJECT_PATH"),
			Branch:     os.Getenv("CI_COMMIT_REF_NAME"),
			Commit:     os.Getenv("CI_COMMIT_SHA"),
			CIProvider: "gitlab",
			CIRunID:    os.Getenv("CI_PIPELINE_ID"),
			RunURL:     os.Getenv("CI_PIPELINE_URL"),
			PRNumber:   os.Getenv("CI_MERGE_REQUEST_IID"),
		}
	}
	return RunMeta{}
}

func FromFile(path string) (RunMeta, error) {
	if strings.TrimSpace(path) == "" {
		return RunMeta{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return RunMeta{}, err
	}
	var out RunMeta
	if err := json.Unmarshal(b, &out); err != nil {
		return RunMeta{}, fmt.Errorf("invalid run-meta json: %w", err)
	}
	return out, nil
}

func Merge(preferFile bool, env RunMeta, file RunMeta) RunMeta {
	if !preferFile {
		return mergeFields(env, file)
	}
	return mergeFields(file, env)
}

func mergeFields(primary RunMeta, fallback RunMeta) RunMeta {
	out := primary
	if out.Repo == "" {
		out.Repo = fallback.Repo
	}
	if out.Branch == "" {
		out.Branch = fallback.Branch
	}
	if out.Commit == "" {
		out.Commit = fallback.Commit
	}
	if out.CIProvider == "" {
		out.CIProvider = fallback.CIProvider
	}
	if out.CIRunID == "" {
		out.CIRunID = fallback.CIRunID
	}
	if out.RunURL == "" {
		out.RunURL = fallback.RunURL
	}
	if out.PRNumber == "" {
		out.PRNumber = fallback.PRNumber
	}
	return out
}

func githubRunURL(repo, runID string) string {
	if strings.TrimSpace(repo) == "" || strings.TrimSpace(runID) == "" {
		return ""
	}
	return "https://github.com/" + repo + "/actions/runs/" + runID
}

func githubPRFromRef(ref string) string {
	parts := strings.Split(strings.TrimSpace(ref), "/")
	if len(parts) >= 3 && parts[0] == "refs" && parts[1] == "pull" {
		return parts[2]
	}
	return ""
}
