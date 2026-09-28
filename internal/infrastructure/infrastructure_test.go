package infrastructure_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"github.com/forgedash/forgedash/internal/infrastructure/auth"
	"github.com/forgedash/forgedash/internal/infrastructure/config"
	"github.com/forgedash/forgedash/internal/infrastructure/git"
	"github.com/forgedash/forgedash/internal/infrastructure/provider/github"
	"github.com/forgedash/forgedash/internal/infrastructure/provider/gitlab"
	"github.com/forgedash/forgedash/internal/infrastructure/provider/mock"
)

func TestGitHubClient_MockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"login":      "octocat",
				"name":       "The Octocat",
				"avatar_url": "https://avatars.githubusercontent.com/u/583231",
			})
		case "/search/issues":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total_count": 1,
				"items": []map[string]any{
					{
						"number":         142,
						"title":          "Test GitHub PR",
						"state":          "open",
						"repository_url": "https://api.github.com/repos/octocat/hello-world",
						"user": map[string]any{
							"login": "octocat",
						},
						"pull_request": map[string]any{},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	acc := account.Account{
		ID:       "gh-test",
		Name:     "GH Test",
		Provider: account.ForgeGitHub,
		Host:     "github.com",
	}
	client := github.NewClient(acc, "fake-token").WithBaseURL(ts.URL)

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error fetching current user: %v", err)
	}
	if user.Username != "octocat" {
		t.Errorf("expected octocat, got %s", user.Username)
	}

	changes, _, err := client.ListChanges(context.Background(), shared.QueryFilter{})
	if err != nil {
		t.Fatalf("unexpected error listing changes: %v", err)
	}
	if len(changes) != 1 || changes[0].Number != 142 {
		t.Fatalf("expected PR 142, got %+v", changes)
	}
}

func TestGitLabClient_MockServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/user":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"username": "tanuki",
				"name":     "Tanuki User",
			})
		case "/merge_requests":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"iid":     88,
					"title":   "Test GitLab MR",
					"state":   "opened",
					"web_url": "https://gitlab.com/gitlab-org/gitlab/-/merge_requests/88",
					"author": map[string]any{
						"username": "tanuki",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	acc := account.Account{
		ID:       "gl-test",
		Name:     "GL Test",
		Provider: account.ForgeGitLab,
		Host:     "gitlab.com",
	}
	client := gitlab.NewClient(acc, "fake-token").WithBaseURL(ts.URL)

	user, err := client.CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "tanuki" {
		t.Errorf("expected tanuki, got %s", user.Username)
	}

	changes, _, err := client.ListChanges(context.Background(), shared.QueryFilter{})
	if err != nil {
		t.Fatalf("unexpected error listing changes: %v", err)
	}
	if len(changes) != 1 || changes[0].Number != 88 {
		t.Fatalf("expected MR 88, got %+v", changes)
	}
}

func TestMockClients_Fixtures(t *testing.T) {
	clients := mock.BuildDemoClients()
	if len(clients) != 3 {
		t.Fatalf("expected 3 demo clients, got %d", len(clients))
	}

	gh := clients[0]
	changes, _, err := gh.ListChanges(context.Background(), shared.QueryFilter{})
	if err != nil || len(changes) == 0 {
		t.Fatalf("expected demo PRs from mock client, got %v (len %d)", err, len(changes))
	}

	diff, err := gh.GetDiff(context.Background(), "charmbracelet", "bubbletea", 142)
	if err != nil || len(diff) == 0 {
		t.Fatalf("expected diff for PR 142, got %v (len %d)", err, len(diff))
	}
}

func TestAuthStore_FileAndEnv(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "credentials.json")

	store, err := auth.NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	if err := store.Set("my-ref", "token-12345"); err != nil {
		t.Fatalf("failed to set credential: %v", err)
	}

	got, err := store.Get("my-ref")
	if err != nil || got != "token-12345" {
		t.Fatalf("expected token-12345, got %s (err %v)", got, err)
	}

	t.Setenv("FORGEDASH_MY_REF_TOKEN", "override-token")
	gotEnv, err := store.Get("my-ref")
	if err != nil || gotEnv != "override-token" {
		t.Fatalf("expected override-token, got %s", gotEnv)
	}
}

func TestConfigLoadAndInit(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "config.yaml")

	if err := config.Init(confPath); err != nil {
		t.Fatalf("failed to init config: %v", err)
	}

	cfg, err := config.Load(confPath)
	if err != nil {
		t.Fatalf("failed to load initial config: %v", err)
	}

	if cfg.DefaultView != "prs" {
		t.Errorf("expected default view prs, got %s", cfg.DefaultView)
	}
	if len(cfg.PRSections) == 0 {
		t.Errorf("expected default PR sections")
	}
}

func TestGitRemoteParsing(t *testing.T) {
	tests := []struct {
		remote    string
		wantOwner string
		wantRepo  string
		wantHost  string
		wantType  account.ForgeType
	}{
		{
			remote:    "git@github.com:charmbracelet/bubbletea.git",
			wantOwner: "charmbracelet",
			wantRepo:  "bubbletea",
			wantHost:  "github.com",
			wantType:  account.ForgeGitHub,
		},
		{
			remote:    "https://gitlab.com/gitlab-org/gitlab-runner.git",
			wantOwner: "gitlab-org",
			wantRepo:  "gitlab-runner",
			wantHost:  "gitlab.com",
			wantType:  account.ForgeGitLab,
		},
		{
			remote:    "git@codeberg.org:forgejo/forgejo.git",
			wantOwner: "forgejo",
			wantRepo:  "forgejo",
			wantHost:  "codeberg.org",
			wantType:  account.ForgeForgejo,
		},
	}

	for _, tc := range tests {
		owner, repo, host, fType, err := git.ParseGitRemote(tc.remote)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", tc.remote, err)
		}
		if owner != tc.wantOwner || repo != tc.wantRepo || host != tc.wantHost || fType != tc.wantType {
			t.Errorf("ParseGitRemote(%s) = (%s, %s, %s, %s); want (%s, %s, %s, %s)",
				tc.remote, owner, repo, host, fType, tc.wantOwner, tc.wantRepo, tc.wantHost, tc.wantType)
		}
	}
}
