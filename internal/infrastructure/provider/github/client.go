package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/forgedash/forgedash/internal/application/ports"
	"github.com/forgedash/forgedash/internal/domain/account"
	"github.com/forgedash/forgedash/internal/domain/change"
	"github.com/forgedash/forgedash/internal/domain/forge"
	"github.com/forgedash/forgedash/internal/domain/issue"
	"github.com/forgedash/forgedash/internal/domain/notification"
	"github.com/forgedash/forgedash/internal/domain/repository"
	"github.com/forgedash/forgedash/internal/domain/review"
	"github.com/forgedash/forgedash/internal/domain/shared"
	"github.com/forgedash/forgedash/internal/infrastructure/httpclient"
)

type Client struct {
	acc        account.Account
	token      string
	baseURL    string
	httpClient *httpclient.Client
}

func NewClient(acc account.Account, token string) *Client {
	baseURL := "https://api.github.com"
	if acc.Host != "" && acc.Host != "github.com" {
		baseURL = fmt.Sprintf("https://%s/api/v3", acc.Host)
	}
	return &Client{
		acc:        acc,
		token:      token,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: httpclient.NewClient(15 * time.Second),
	}
}

func (c *Client) WithBaseURL(u string) *Client {
	c.baseURL = strings.TrimSuffix(u, "/")
	return c
}

func (c *Client) Account() account.Account {
	return c.acc
}

func (c *Client) Capabilities() forge.CapabilityMatrix {
	return forge.GitHubCapabilities()
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
		bodyReader = buf
	}

	endpoint := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		endpoint = fmt.Sprintf("%s%s", c.baseURL, path)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) CurrentUser(ctx context.Context) (shared.User, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/user", nil)
	if err != nil {
		return shared.User{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return shared.User{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var dto struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		HTMLURL   string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return shared.User{}, err
	}
	return shared.User{
		Username:    dto.Login,
		DisplayName: dto.Name,
		AvatarURL:   dto.AvatarURL,
		HTMLURL:     dto.HTMLURL,
	}, nil
}

func (c *Client) ListRepositories(ctx context.Context, filter shared.QueryFilter) ([]repository.Repository, shared.Pagination, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/user/repos?per_page=50&sort=updated", nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var repos []struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		Private       bool   `json:"private"`
		Fork          bool   `json:"fork"`
		DefaultBranch string `json:"default_branch"`
		HTMLURL       string `json:"html_url"`
		SSHURL        string `json:"ssh_url"`
		CloneURL      string `json:"clone_url"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []repository.Repository
	for _, r := range repos {
		results = append(results, repository.Repository{
			ID:            fmt.Sprintf("%d", r.ID),
			AccountID:     c.acc.ID,
			Provider:      account.ForgeGitHub,
			Host:          c.acc.Host,
			Owner:         r.Owner.Login,
			Name:          r.Name,
			FullName:      r.FullName,
			Description:   r.Description,
			DefaultBranch: r.DefaultBranch,
			IsPrivate:     r.Private,
			IsFork:        r.Fork,
			HTMLURL:       r.HTMLURL,
			SSHURL:        r.SSHURL,
			CloneURL:      r.CloneURL,
		})
	}
	return results, shared.Pagination{Page: 1, PerPage: len(results)}, nil
}

func (c *Client) ListChanges(ctx context.Context, filter shared.QueryFilter) ([]change.ChangeRequest, shared.Pagination, error) {
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 30
	}

	searchQuery := "is:pr is:open"
	if filter.Query != "" {
		searchQuery = filter.Query
		if !strings.Contains(searchQuery, "is:pr") {
			searchQuery = "is:pr " + searchQuery
		}
	}
	if filter.Repo != "" {
		searchQuery = fmt.Sprintf("repo:%s %s", filter.Repo, searchQuery)
	}

	path := fmt.Sprintf("/search/issues?q=%s&per_page=%d&sort=updated", url.QueryEscape(searchQuery), perPage)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var searchResp struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			Number        int       `json:"number"`
			Title         string    `json:"title"`
			State         string    `json:"state"`
			HTMLURL       string    `json:"html_url"`
			Body          string    `json:"body"`
			Comments      int       `json:"comments"`
			CreatedAt     time.Time `json:"created_at"`
			UpdatedAt     time.Time `json:"updated_at"`
			RepositoryURL string    `json:"repository_url"`
			User          struct {
				Login     string `json:"login"`
				AvatarURL string `json:"avatar_url"`
				HTMLURL   string `json:"html_url"`
			} `json:"user"`
			Labels []struct {
				Name        string `json:"name"`
				Color       string `json:"color"`
				Description string `json:"description"`
			} `json:"labels"`
			PullRequest struct {
				MergedAt *time.Time `json:"merged_at"`
				DiffURL  string     `json:"diff_url"`
			} `json:"pull_request"`
			Draft bool `json:"draft"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []change.ChangeRequest
	for _, item := range searchResp.Items {
		owner, repoName := parseRepoURL(item.RepositoryURL)

		st := change.ChangeStateOpen
		if item.PullRequest.MergedAt != nil {
			st = change.ChangeStateMerged
		} else if item.State == "closed" {
			st = change.ChangeStateClosed
		} else if item.Draft {
			st = change.ChangeStateDraft
		}

		var labels []shared.Label
		for _, l := range item.Labels {
			labels = append(labels, shared.Label{
				Name:        l.Name,
				Color:       l.Color,
				Description: l.Description,
			})
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeGitHub,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repoName,
			ResourceType: forge.ResourceTypePR,
			NativeID:     fmt.Sprintf("%d", item.Number),
		}

		results = append(results, change.ChangeRequest{
			ResourceID: resID,
			AccountID:  c.acc.ID,
			Number:     item.Number,
			Title:      item.Title,
			State:      st,
			RepoOwner:  owner,
			RepoName:   repoName,
			Author: shared.User{
				Username:  item.User.Login,
				AvatarURL: item.User.AvatarURL,
				HTMLURL:   item.User.HTMLURL,
			},
			Body:         item.Body,
			Labels:       labels,
			Comments:     item.Comments,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
			HTMLURL:      item.HTMLURL,
			DiffURL:      item.PullRequest.DiffURL,
			CIStatus:     change.CIStatusPassing,
			ReviewStatus: change.ReviewStatusNone,
		})
	}

	return results, shared.Pagination{
		Page:        1,
		PerPage:     perPage,
		Total:       searchResp.TotalCount,
		HasNextPage: len(results) == perPage,
	}, nil
}

func (c *Client) GetChange(ctx context.Context, owner, repo string, number int) (*change.ChangeRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var dto struct {
		Number       int        `json:"number"`
		Title        string     `json:"title"`
		State        string     `json:"state"`
		Body         string     `json:"body"`
		CreatedAt    time.Time  `json:"created_at"`
		UpdatedAt    time.Time  `json:"updated_at"`
		ClosedAt     *time.Time `json:"closed_at"`
		MergedAt     *time.Time `json:"merged_at"`
		HTMLURL      string     `json:"html_url"`
		DiffURL      string     `json:"diff_url"`
		Draft        bool       `json:"draft"`
		Additions    int        `json:"additions"`
		Deletions    int        `json:"deletions"`
		ChangedFiles int        `json:"changed_files"`
		Head         struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
		User struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
			HTMLURL   string `json:"html_url"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	st := change.ChangeStateOpen
	if dto.MergedAt != nil {
		st = change.ChangeStateMerged
	} else if dto.State == "closed" {
		st = change.ChangeStateClosed
	} else if dto.Draft {
		st = change.ChangeStateDraft
	}

	cr := &change.ChangeRequest{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitHub,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypePR,
			NativeID:     fmt.Sprintf("%d", dto.Number),
		},
		AccountID:    c.acc.ID,
		Number:       dto.Number,
		Title:        dto.Title,
		State:        st,
		RepoOwner:    owner,
		RepoName:     repo,
		SourceBranch: dto.Head.Ref,
		TargetBranch: dto.Base.Ref,
		Author: shared.User{
			Username:  dto.User.Login,
			AvatarURL: dto.User.AvatarURL,
			HTMLURL:   dto.User.HTMLURL,
		},
		Body:      dto.Body,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		ClosedAt:  dto.ClosedAt,
		MergedAt:  dto.MergedAt,
		HTMLURL:   dto.HTMLURL,
		DiffURL:   dto.DiffURL,
		Stats: change.DiffStats{
			Additions: dto.Additions,
			Deletions: dto.Deletions,
			Files:     dto.ChangedFiles,
		},
	}
	return cr, nil
}

func (c *Client) GetDiff(ctx context.Context, owner, repo string, number int) ([]shared.DiffFile, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/files?per_page=100", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var files []struct {
		Filename  string `json:"filename"`
		Status    string `json:"status"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
		Patch     string `json:"patch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	var results []shared.DiffFile
	for _, f := range files {
		var st shared.FileStatus
		switch f.Status {
		case "added":
			st = shared.FileStatusAdded
		case "removed":
			st = shared.FileStatusDeleted
		case "modified":
			st = shared.FileStatusModified
		default:
			st = shared.FileStatusModified
		}
		results = append(results, shared.DiffFile{
			NewPath:   f.Filename,
			Status:    st,
			Additions: f.Additions,
			Deletions: f.Deletions,
			Patch:     f.Patch,
		})
	}
	return results, nil
}

func (c *Client) ListCommits(ctx context.Context, owner, repo string, number int) ([]shared.Commit, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/commits?per_page=100", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var commits []struct {
		SHA     string `json:"sha"`
		HTMLURL string `json:"html_url"`
		Commit  struct {
			Message string `json:"message"`
			Author  struct {
				Name string    `json:"name"`
				Date time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	var results []shared.Commit
	for _, cm := range commits {
		shortSHA := cm.SHA
		if len(shortSHA) > 7 {
			shortSHA = shortSHA[:7]
		}
		results = append(results, shared.Commit{
			SHA:      cm.SHA,
			ShortSHA: shortSHA,
			Author: shared.User{
				Username:    cm.Author.Login,
				DisplayName: cm.Commit.Author.Name,
			},
			Message: strings.Split(cm.Commit.Message, "\n")[0],
			Date:    cm.Commit.Author.Date,
			HTMLURL: cm.HTMLURL,
		})
	}
	return results, nil
}

func (c *Client) ListChecks(ctx context.Context, owner, repo string, ref string) ([]shared.CheckRun, error) {
	path := fmt.Sprintf("/repos/%s/%s/commits/%s/check-runs", owner, repo, ref)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var checkResp struct {
		CheckRuns []struct {
			ID          int64      `json:"id"`
			Name        string     `json:"name"`
			Status      string     `json:"status"`
			Conclusion  *string    `json:"conclusion"`
			HTMLURL     string     `json:"html_url"`
			StartedAt   *time.Time `json:"started_at"`
			CompletedAt *time.Time `json:"completed_at"`
		} `json:"check_runs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
		return nil, err
	}

	var results []shared.CheckRun
	for _, cr := range checkResp.CheckRuns {
		var conc shared.CheckConclusion
		if cr.Conclusion != nil {
			switch *cr.Conclusion {
			case "success":
				conc = shared.CheckConclusionSuccess
			case "failure":
				conc = shared.CheckConclusionFailure
			case "cancelled":
				conc = shared.CheckConclusionCancelled
			default:
				conc = shared.CheckConclusionNeutral
			}
		} else {
			conc = shared.CheckConclusionPending
		}

		var st shared.CheckStatus
		switch cr.Status {
		case "completed":
			st = shared.CheckStatusCompleted
		case "in_progress":
			st = shared.CheckStatusInProgress
		default:
			st = shared.CheckStatusQueued
		}

		results = append(results, shared.CheckRun{
			ID:          fmt.Sprintf("%d", cr.ID),
			Name:        cr.Name,
			Status:      st,
			Conclusion:  conc,
			URL:         cr.HTMLURL,
			StartedAt:   cr.StartedAt,
			CompletedAt: cr.CompletedAt,
		})
	}
	return results, nil
}

func (c *Client) ListComments(ctx context.Context, owner, repo string, number int, isChange bool) ([]shared.Comment, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var comments []struct {
		ID        int64     `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		HTMLURL   string    `json:"html_url"`
		User      struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
			HTMLURL   string `json:"html_url"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	var results []shared.Comment
	for _, cm := range comments {
		results = append(results, shared.Comment{
			ID: fmt.Sprintf("%d", cm.ID),
			Author: shared.User{
				Username:  cm.User.Login,
				AvatarURL: cm.User.AvatarURL,
				HTMLURL:   cm.User.HTMLURL,
			},
			Body:      cm.Body,
			CreatedAt: cm.CreatedAt,
			HTMLURL:   cm.HTMLURL,
		})
	}
	return results, nil
}

func (c *Client) CreateComment(ctx context.Context, owner, repo string, number int, body string, isChange bool) (*shared.Comment, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodPost, path, map[string]string{"body": body})
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var cm struct {
		ID        int64     `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		HTMLURL   string    `json:"html_url"`
		User      struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
			HTMLURL   string `json:"html_url"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cm); err != nil {
		return nil, err
	}

	return &shared.Comment{
		ID: fmt.Sprintf("%d", cm.ID),
		Author: shared.User{
			Username:  cm.User.Login,
			AvatarURL: cm.User.AvatarURL,
			HTMLURL:   cm.User.HTMLURL,
		},
		Body:      cm.Body,
		CreatedAt: cm.CreatedAt,
		HTMLURL:   cm.HTMLURL,
	}, nil
}

func (c *Client) ListReviews(ctx context.Context, owner, repo string, number int) ([]review.Review, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var reviews []struct {
		ID          int64     `json:"id"`
		State       string    `json:"state"`
		Body        string    `json:"body"`
		SubmittedAt time.Time `json:"submitted_at"`
		HTMLURL     string    `json:"html_url"`
		User        struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
			HTMLURL   string `json:"html_url"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&reviews); err != nil {
		return nil, err
	}

	var results []review.Review
	for _, r := range reviews {
		var st review.ReviewState
		switch r.State {
		case "APPROVED":
			st = review.ReviewStateApproved
		case "CHANGES_REQUESTED":
			st = review.ReviewStateChangesRequested
		case "COMMENTED":
			st = review.ReviewStateCommented
		default:
			st = review.ReviewStatePending
		}
		results = append(results, review.Review{
			ID: fmt.Sprintf("%d", r.ID),
			Author: shared.User{
				Username:  r.User.Login,
				AvatarURL: r.User.AvatarURL,
				HTMLURL:   r.User.HTMLURL,
			},
			State:       st,
			Body:        r.Body,
			SubmittedAt: r.SubmittedAt,
			HTMLURL:     r.HTMLURL,
		})
	}
	return results, nil
}

func (c *Client) SubmitReview(ctx context.Context, owner, repo string, number int, state review.ReviewState, body string) error {
	event := "COMMENT"
	switch state {
	case review.ReviewStateApproved:
		event = "APPROVE"
	case review.ReviewStateChangesRequested:
		event = "REQUEST_CHANGES"
	}

	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodPost, path, map[string]string{
		"event": event,
		"body":  body,
	})
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer httpclient.DrainAndClose(resp.Body)
	return nil
}

func (c *Client) MergeChange(ctx context.Context, owner, repo string, number int, method string, title, message string) error {
	if method == "" {
		method = "merge"
	}
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, number)
	payload := map[string]string{
		"merge_method": method,
	}
	if title != "" {
		payload["commit_title"] = title
	}
	if message != "" {
		payload["commit_message"] = message
	}

	req, err := c.newRequest(ctx, http.MethodPut, path, payload)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer httpclient.DrainAndClose(resp.Body)
	return nil
}

func (c *Client) CloseChange(ctx context.Context, owner, repo string, number int) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodPatch, path, map[string]string{"state": "closed"})
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer httpclient.DrainAndClose(resp.Body)
	return nil
}

func (c *Client) ReopenChange(ctx context.Context, owner, repo string, number int) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodPatch, path, map[string]string{"state": "open"})
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer httpclient.DrainAndClose(resp.Body)
	return nil
}

func (c *Client) ListIssues(ctx context.Context, filter shared.QueryFilter) ([]issue.Issue, shared.Pagination, error) {
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 30
	}

	searchQuery := "is:issue is:open"
	if filter.Query != "" {
		searchQuery = filter.Query
		if !strings.Contains(searchQuery, "is:issue") {
			searchQuery = "is:issue " + searchQuery
		}
	}
	if filter.Repo != "" {
		searchQuery = fmt.Sprintf("repo:%s %s", filter.Repo, searchQuery)
	}

	path := fmt.Sprintf("/search/issues?q=%s&per_page=%d&sort=updated", url.QueryEscape(searchQuery), perPage)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var searchResp struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			Number        int       `json:"number"`
			Title         string    `json:"title"`
			State         string    `json:"state"`
			HTMLURL       string    `json:"html_url"`
			Body          string    `json:"body"`
			Comments      int       `json:"comments"`
			CreatedAt     time.Time `json:"created_at"`
			UpdatedAt     time.Time `json:"updated_at"`
			RepositoryURL string    `json:"repository_url"`
			User          struct {
				Login     string `json:"login"`
				AvatarURL string `json:"avatar_url"`
				HTMLURL   string `json:"html_url"`
			} `json:"user"`
			Labels []struct {
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"labels"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []issue.Issue
	for _, item := range searchResp.Items {
		owner, repoName := parseRepoURL(item.RepositoryURL)

		var labels []shared.Label
		for _, l := range item.Labels {
			labels = append(labels, shared.Label{Name: l.Name, Color: l.Color})
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeGitHub,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repoName,
			ResourceType: forge.ResourceTypeIssue,
			NativeID:     fmt.Sprintf("%d", item.Number),
		}

		st := issue.IssueStateOpen
		if item.State == "closed" {
			st = issue.IssueStateClosed
		}

		results = append(results, issue.Issue{
			ResourceID: resID,
			AccountID:  c.acc.ID,
			Number:     item.Number,
			Title:      item.Title,
			State:      st,
			RepoOwner:  owner,
			RepoName:   repoName,
			Author: shared.User{
				Username:  item.User.Login,
				AvatarURL: item.User.AvatarURL,
				HTMLURL:   item.User.HTMLURL,
			},
			Body:      item.Body,
			Labels:    labels,
			Comments:  item.Comments,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			HTMLURL:   item.HTMLURL,
		})
	}

	return results, shared.Pagination{
		Page:        1,
		PerPage:     perPage,
		Total:       searchResp.TotalCount,
		HasNextPage: len(results) == perPage,
	}, nil
}

func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*issue.Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var dto struct {
		Number    int        `json:"number"`
		Title     string     `json:"title"`
		State     string     `json:"state"`
		Body      string     `json:"body"`
		Comments  int        `json:"comments"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		ClosedAt  *time.Time `json:"closed_at"`
		HTMLURL   string     `json:"html_url"`
		User      struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
			HTMLURL   string `json:"html_url"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	st := issue.IssueStateOpen
	if dto.State == "closed" {
		st = issue.IssueStateClosed
	}

	return &issue.Issue{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitHub,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypeIssue,
			NativeID:     fmt.Sprintf("%d", dto.Number),
		},
		AccountID: c.acc.ID,
		Number:    dto.Number,
		Title:     dto.Title,
		State:     st,
		RepoOwner: owner,
		RepoName:  repo,
		Author: shared.User{
			Username:  dto.User.Login,
			AvatarURL: dto.User.AvatarURL,
			HTMLURL:   dto.User.HTMLURL,
		},
		Body:      dto.Body,
		Comments:  dto.Comments,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		ClosedAt:  dto.ClosedAt,
		HTMLURL:   dto.HTMLURL,
	}, nil
}

func (c *Client) CloseIssue(ctx context.Context, owner, repo string, number int) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodPatch, path, map[string]string{"state": "closed"})
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer httpclient.DrainAndClose(resp.Body)
	return nil
}

func (c *Client) ReopenIssue(ctx context.Context, owner, repo string, number int) error {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodPatch, path, map[string]string{"state": "open"})
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer httpclient.DrainAndClose(resp.Body)
	return nil
}

func (c *Client) ListNotifications(ctx context.Context, filter shared.QueryFilter) ([]notification.Notification, shared.Pagination, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/notifications?per_page=50", nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var items []struct {
		ID         string    `json:"id"`
		Unread     bool      `json:"unread"`
		Reason     string    `json:"reason"`
		UpdatedAt  time.Time `json:"updated_at"`
		Repository struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"repository"`
		Subject struct {
			Title string `json:"title"`
			URL   string `json:"url"`
			Type  string `json:"type"`
		} `json:"subject"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []notification.Notification
	for _, it := range items {
		var subType notification.SubjectType
		switch it.Subject.Type {
		case "PullRequest":
			subType = notification.SubjectTypePR
		case "Issue":
			subType = notification.SubjectTypeIssue
		default:
			subType = notification.SubjectTypeCommit
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeGitHub,
			Host:         c.acc.Host,
			RepoOwner:    it.Repository.Owner.Login,
			RepoName:     it.Repository.Name,
			ResourceType: forge.ResourceTypeNotification,
			NativeID:     it.ID,
		}

		results = append(results, notification.Notification{
			ID:           it.ID,
			AccountID:    c.acc.ID,
			ResourceID:   resID,
			RepoOwner:    it.Repository.Owner.Login,
			RepoName:     it.Repository.Name,
			SubjectTitle: it.Subject.Title,
			SubjectType:  subType,
			Reason:       it.Reason,
			Unread:       it.Unread,
			UpdatedAt:    it.UpdatedAt,
			HTMLURL:      fmt.Sprintf("https://%s/%s/%s", c.acc.Host, it.Repository.Owner.Login, it.Repository.Name),
		})
	}
	return results, shared.Pagination{Page: 1, PerPage: len(results)}, nil
}

func parseRepoURL(raw string) (string, string) {

	parts := strings.Split(strings.TrimSuffix(raw, "/"), "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}
	return "unknown", "unknown"
}

var _ ports.ForgeClient = (*Client)(nil)
