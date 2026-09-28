package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/forgedash/forgedash/internal/domain/account"
)

type Service struct {
	workDir string
}

func NewService(workDir string) *Service {
	return &Service{workDir: workDir}
}

func (s *Service) DetectRepo(ctx context.Context, dir string) (owner, repo, host string, forgeType account.ForgeType, err error) {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", "", "", "", fmt.Errorf("not a git repository or no origin remote: %w", err)
	}

	remoteURL := strings.TrimSpace(string(out))
	return ParseGitRemote(remoteURL)
}

func ParseGitRemote(remote string) (owner, repo, host string, forgeType account.ForgeType, err error) {
	clean := strings.TrimSuffix(remote, ".git")

	if strings.HasPrefix(clean, "git@") {
		parts := strings.Split(strings.TrimPrefix(clean, "git@"), ":")
		if len(parts) == 2 {
			host = parts[0]
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 2 {
				owner = strings.Join(pathParts[:len(pathParts)-1], "/")
				repo = pathParts[len(pathParts)-1]
				forgeType = inferForgeType(host)
				return owner, repo, host, forgeType, nil
			}
		}
	}

	if strings.HasPrefix(clean, "https://") || strings.HasPrefix(clean, "http://") {
		trimmed := strings.TrimPrefix(strings.TrimPrefix(clean, "https://"), "http://")
		parts := strings.Split(trimmed, "/")
		if len(parts) >= 3 {
			host = parts[0]
			owner = strings.Join(parts[1:len(parts)-1], "/")
			repo = parts[len(parts)-1]
			forgeType = inferForgeType(host)
			return owner, repo, host, forgeType, nil
		}
	}

	return "", "", "", "", fmt.Errorf("unrecognized git remote URL: %s", remote)
}

func inferForgeType(host string) account.ForgeType {
	switch {
	case strings.Contains(host, "gitlab"):
		return account.ForgeGitLab
	case strings.Contains(host, "codeberg") || strings.Contains(host, "forgejo") || strings.Contains(host, "gitea"):
		return account.ForgeForgejo
	default:
		return account.ForgeGitHub
	}
}

func (s *Service) CheckoutBranch(ctx context.Context, branch string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", branch)
	if s.workDir != "" {
		cmd.Dir = s.workDir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %s (%w)", string(out), err)
	}
	return nil
}

func (s *Service) FetchAndCheckoutPR(ctx context.Context, prNumber int, branchName string) error {
	fetchRef := fmt.Sprintf("pull/%d/head:%s", prNumber, branchName)
	cmd := exec.CommandContext(ctx, "git", "fetch", "origin", fetchRef)
	if s.workDir != "" {
		cmd.Dir = s.workDir
	}
	if out, err := cmd.CombinedOutput(); err != nil {

		glRef := fmt.Sprintf("merge-requests/%d/head:%s", prNumber, branchName)
		cmdGL := exec.CommandContext(ctx, "git", "fetch", "origin", glRef)
		if s.workDir != "" {
			cmdGL.Dir = s.workDir
		}
		if outGL, errGL := cmdGL.CombinedOutput(); errGL != nil {
			return fmt.Errorf("git fetch PR failed: %s / %s (%w)", string(out), string(outGL), err)
		}
	}

	return s.CheckoutBranch(ctx, branchName)
}

func (s *Service) HasDelta() bool {
	_, err := exec.LookPath("delta")
	return err == nil
}

func (s *Service) RenderDiffWithDelta(diffContent string) (string, error) {
	if !s.HasDelta() {
		return diffContent, nil
	}

	cmd := exec.Command("delta", "--paging=never", "--decorations=never")
	cmd.Stdin = strings.NewReader(diffContent)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return diffContent, err
	}
	return out.String(), nil
}
