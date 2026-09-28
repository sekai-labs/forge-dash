package domain_test

import (
	"testing"

	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/issue"
)

func TestAccountValidation(t *testing.T) {
	acc := account.Account{
		ID:       "gh-main",
		Name:     "GitHub Main",
		Provider: account.ForgeGitHub,
		Host:     "github.com",
	}
	if err := acc.Validate(); err != nil {
		t.Fatalf("expected valid account, got: %v", err)
	}

	invalid := account.Account{
		ID:       "",
		Name:     "No ID",
		Provider: "unknown",
	}
	if err := invalid.Validate(); err == nil {
		t.Fatalf("expected error for invalid account, got nil")
	}
}

func TestResourceID_RoundTrip(t *testing.T) {
	orig := forge.ResourceID{
		Provider:     account.ForgeGitHub,
		Host:         "github.com",
		RepoOwner:    "charmbracelet",
		RepoName:     "bubbletea",
		ResourceType: forge.ResourceTypePR,
		NativeID:     "123",
	}

	uri := orig.URI()
	expected := "forgedash://github/github.com/charmbracelet/bubbletea/pull/123"
	if uri != expected {
		t.Fatalf("expected uri %s, got %s", expected, uri)
	}

	parsed, err := forge.ParseResourceID(uri)
	if err != nil {
		t.Fatalf("failed to parse resource ID: %v", err)
	}
	if parsed != orig {
		t.Fatalf("expected %+v, got %+v", orig, parsed)
	}
}

func TestCapabilityMatrix(t *testing.T) {
	ghCaps := forge.GitHubCapabilities()
	if !ghCaps.Has(forge.CapPullRequests) {
		t.Errorf("GitHub should support pull requests")
	}
	if !ghCaps.Has(forge.CapCheckRuns) {
		t.Errorf("GitHub should support check runs")
	}

	glCaps := forge.GitLabCapabilities()
	if !glCaps.Has(forge.CapPipelines) {
		t.Errorf("GitLab should support pipelines")
	}
}

func TestChangeRequestFormatting(t *testing.T) {
	cr := change.ChangeRequest{
		Number:    42,
		RepoOwner: "forgedash",
		RepoName:  "forgedash",
		ResourceID: forge.ResourceID{
			ResourceType: forge.ResourceTypePR,
		},
	}
	if cr.FullRepoName() != "forgedash/forgedash" {
		t.Errorf("expected full repo name forgedash/forgedash, got %s", cr.FullRepoName())
	}
	if cr.DisplayID() != "#42" {
		t.Errorf("expected #42, got %s", cr.DisplayID())
	}

	mr := change.ChangeRequest{
		Number: 99,
		ResourceID: forge.ResourceID{
			ResourceType: forge.ResourceTypeMR,
		},
	}
	if mr.DisplayID() != "!99" {
		t.Errorf("expected !99, got %s", mr.DisplayID())
	}
}

func TestIssueFormatting(t *testing.T) {
	iss := issue.Issue{
		Number:    101,
		RepoOwner: "forgedash",
		RepoName:  "core",
	}
	if iss.FullRepoName() != "forgedash/core" {
		t.Errorf("expected forgedash/core, got %s", iss.FullRepoName())
	}
	if iss.DisplayID() != "#101" {
		t.Errorf("expected #101, got %s", iss.DisplayID())
	}
}
