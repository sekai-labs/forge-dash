package forgejo

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
	baseURL := "https://codeberg.org/api/v1"
	if acc.Host != "" && acc.Host != "codeberg.org" {
		baseURL = fmt.Sprintf("https://%s/api/v1", acc.Host)
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

func (c *Client) Account() account.Account { return c.acc }

func (c *Client) Capabilities() forge.CapabilityMatrix {
	return forge.ForgejoCapabilities()
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

	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
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
		Username  string `json:"username"`
		FullName  string `json:"full_name"`
		AvatarURL string `json:"avatar_url"`
		HTMLURL   string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return shared.User{}, err
	}

	return shared.User{
		Username:    dto.Username,
		DisplayName: dto.FullName,
		AvatarURL:   dto.AvatarURL,
		HTMLURL:     dto.HTMLURL,
	}, nil
}

func (c *Client) ListRepositories(ctx context.Context, filter shared.QueryFilter) ([]repository.Repository, shared.Pagination, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/user/repos?limit=50", nil)
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
			Username string `json:"username"`
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
			Provider:      account.ForgeForgejo,
			Host:          c.acc.Host,
			Owner:         r.Owner.Username,
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

	path := fmt.Sprintf("/repos/issues/search?type=pulls&limit=%d&sort=recentupdate", perPage)
	if filter.State == "closed" {
		path += "&state=closed"
	} else {
		path += "&state=open"
	}
	if filter.Query != "" {
		path += fmt.Sprintf("&q=%s", url.QueryEscape(filter.Query))
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var items []struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		State     string    `json:"state"`
		Body      string    `json:"body"`
		HTMLURL   string    `json:"html_url"`
		Comments  int       `json:"comments"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		User      struct {
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
		Repository struct {
			Name  string `json:"name"`
			Owner struct {
				Username string `json:"username"`
			} `json:"owner"`
		} `json:"repository"`
		PullRequest *struct {
			Merged   bool       `json:"merged"`
			MergedAt *time.Time `json:"merged_at"`
			Head     struct {
				Ref string `json:"ref"`
			} `json:"head"`
			Base struct {
				Ref string `json:"ref"`
			} `json:"base"`
		} `json:"pull_request"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []change.ChangeRequest
	for _, it := range items {
		owner := it.Repository.Owner.Username
		repo := it.Repository.Name

		st := change.ChangeStateOpen
		if it.PullRequest != nil && it.PullRequest.Merged {
			st = change.ChangeStateMerged
		} else if it.State == "closed" {
			st = change.ChangeStateClosed
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeForgejo,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypePR,
			NativeID:     fmt.Sprintf("%d", it.Number),
		}

		results = append(results, change.ChangeRequest{
			ResourceID: resID,
			AccountID:  c.acc.ID,
			Number:     it.Number,
			Title:      it.Title,
			State:      st,
			RepoOwner:  owner,
			RepoName:   repo,
			Author: shared.User{
				Username:  it.User.Username,
				AvatarURL: it.User.AvatarURL,
			},
			Body:         it.Body,
			Comments:     it.Comments,
			CreatedAt:    it.CreatedAt,
			UpdatedAt:    it.UpdatedAt,
			HTMLURL:      it.HTMLURL,
			CIStatus:     change.CIStatusPassing,
			ReviewStatus: change.ReviewStatusNone,
		})
	}

	return results, shared.Pagination{Page: 1, PerPage: perPage, HasNextPage: len(results) == perPage}, nil
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

	var pr struct {
		Number    int        `json:"number"`
		Title     string     `json:"title"`
		State     string     `json:"state"`
		Body      string     `json:"body"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		ClosedAt  *time.Time `json:"closed_at"`
		MergedAt  *time.Time `json:"merged_at"`
		HTMLURL   string     `json:"html_url"`
		DiffURL   string     `json:"diff_url"`
		Merged    bool       `json:"merged"`
		Head      struct {
			Ref string `json:"ref"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
		User struct {
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	st := change.ChangeStateOpen
	if pr.Merged {
		st = change.ChangeStateMerged
	} else if pr.State == "closed" {
		st = change.ChangeStateClosed
	}

	return &change.ChangeRequest{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeForgejo,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypePR,
			NativeID:     fmt.Sprintf("%d", pr.Number),
		},
		AccountID:    c.acc.ID,
		Number:       pr.Number,
		Title:        pr.Title,
		State:        st,
		RepoOwner:    owner,
		RepoName:     repo,
		SourceBranch: pr.Head.Ref,
		TargetBranch: pr.Base.Ref,
		Author: shared.User{
			Username:  pr.User.Username,
			AvatarURL: pr.User.AvatarURL,
		},
		Body:      pr.Body,
		CreatedAt: pr.CreatedAt,
		UpdatedAt: pr.UpdatedAt,
		ClosedAt:  pr.ClosedAt,
		MergedAt:  pr.MergedAt,
		HTMLURL:   pr.HTMLURL,
		DiffURL:   pr.DiffURL,
	}, nil
}

func (c *Client) GetDiff(ctx context.Context, owner, repo string, number int) ([]shared.DiffFile, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d.diff", owner, repo, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return []shared.DiffFile{
		{
			NewPath: fmt.Sprintf("PR #%d Diff", number),
			Status:  shared.FileStatusModified,
			Patch:   string(data),
		},
	}, nil
}

func (c *Client) ListCommits(ctx context.Context, owner, repo string, number int) ([]shared.Commit, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/commits", owner, repo, number)
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
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string    `json:"name"`
				Date time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	var results []shared.Commit
	for _, cm := range commits {
		short := cm.SHA
		if len(short) > 7 {
			short = short[:7]
		}
		results = append(results, shared.Commit{
			SHA:      cm.SHA,
			ShortSHA: short,
			Author: shared.User{
				Username:    cm.Commit.Author.Name,
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
	path := fmt.Sprintf("/repos/%s/%s/commits/%s/statuses", owner, repo, ref)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var statuses []struct {
		ID          int64     `json:"id"`
		Context     string    `json:"context"`
		Status      string    `json:"status"`
		Description string    `json:"description"`
		TargetURL   string    `json:"target_url"`
		CreatedAt   time.Time `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
		return nil, err
	}

	var results []shared.CheckRun
	for _, s := range statuses {
		var conc shared.CheckConclusion
		switch s.Status {
		case "success":
			conc = shared.CheckConclusionSuccess
		case "failure", "error":
			conc = shared.CheckConclusionFailure
		case "warning":
			conc = shared.CheckConclusionNeutral
		default:
			conc = shared.CheckConclusionPending
		}

		results = append(results, shared.CheckRun{
			ID:         fmt.Sprintf("%d", s.ID),
			Name:       s.Context,
			Status:     shared.CheckStatusCompleted,
			Conclusion: conc,
			URL:        s.TargetURL,
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
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
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
				Username:  cm.User.Username,
				AvatarURL: cm.User.AvatarURL,
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
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cm); err != nil {
		return nil, err
	}

	return &shared.Comment{
		ID: fmt.Sprintf("%d", cm.ID),
		Author: shared.User{
			Username:  cm.User.Username,
			AvatarURL: cm.User.AvatarURL,
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
		User        struct {
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
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
		case "REQUEST_CHANGES":
			st = review.ReviewStateChangesRequested
		default:
			st = review.ReviewStateCommented
		}
		results = append(results, review.Review{
			ID: fmt.Sprintf("%d", r.ID),
			Author: shared.User{
				Username:  r.User.Username,
				AvatarURL: r.User.AvatarURL,
			},
			State:       st,
			Body:        r.Body,
			SubmittedAt: r.SubmittedAt,
		})
	}
	return results, nil
}

func (c *Client) SubmitReview(ctx context.Context, owner, repo string, number int, state review.ReviewState, body string) error {
	event := "COMMENT"
	if state == review.ReviewStateApproved {
		event = "APPROVED"
	} else if state == review.ReviewStateChangesRequested {
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
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/merge", owner, repo, number)
	if method == "" {
		method = "merge"
	}
	req, err := c.newRequest(ctx, http.MethodPost, path, map[string]string{
		"Do": method,
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

	path := fmt.Sprintf("/repos/issues/search?type=issues&limit=%d&sort=recentupdate", perPage)
	if filter.State == "closed" {
		path += "&state=closed"
	} else {
		path += "&state=open"
	}
	if filter.Query != "" {
		path += fmt.Sprintf("&q=%s", url.QueryEscape(filter.Query))
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var items []struct {
		Number    int       `json:"number"`
		Title     string    `json:"title"`
		State     string    `json:"state"`
		Body      string    `json:"body"`
		HTMLURL   string    `json:"html_url"`
		Comments  int       `json:"comments"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		User      struct {
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
		Repository struct {
			Name  string `json:"name"`
			Owner struct {
				Username string `json:"username"`
			} `json:"owner"`
		} `json:"repository"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []issue.Issue
	for _, it := range items {
		owner := it.Repository.Owner.Username
		repo := it.Repository.Name

		st := issue.IssueStateOpen
		if it.State == "closed" {
			st = issue.IssueStateClosed
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeForgejo,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypeIssue,
			NativeID:     fmt.Sprintf("%d", it.Number),
		}

		results = append(results, issue.Issue{
			ResourceID: resID,
			AccountID:  c.acc.ID,
			Number:     it.Number,
			Title:      it.Title,
			State:      st,
			RepoOwner:  owner,
			RepoName:   repo,
			Author: shared.User{
				Username:  it.User.Username,
				AvatarURL: it.User.AvatarURL,
			},
			Body:      it.Body,
			Comments:  it.Comments,
			CreatedAt: it.CreatedAt,
			UpdatedAt: it.UpdatedAt,
			HTMLURL:   it.HTMLURL,
		})
	}

	return results, shared.Pagination{Page: 1, PerPage: perPage, HasNextPage: len(results) == perPage}, nil
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
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
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
			Provider:     account.ForgeForgejo,
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
			Username:  dto.User.Username,
			AvatarURL: dto.User.AvatarURL,
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
	req, err := c.newRequest(ctx, http.MethodGet, "/notifications?status-types=unread", nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var items []struct {
		ID         int64     `json:"id"`
		Unread     bool      `json:"unread"`
		UpdatedAt  time.Time `json:"updated_at"`
		Repository struct {
			Name  string `json:"name"`
			Owner struct {
				Username string `json:"username"`
			} `json:"owner"`
		} `json:"repository"`
		Subject struct {
			Title string `json:"title"`
			Type  string `json:"type"`
			URL   string `json:"url"`
		} `json:"subject"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []notification.Notification
	for _, it := range items {
		subType := notification.SubjectTypePR
		if it.Subject.Type == "Issue" {
			subType = notification.SubjectTypeIssue
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeForgejo,
			Host:         c.acc.Host,
			RepoOwner:    it.Repository.Owner.Username,
			RepoName:     it.Repository.Name,
			ResourceType: forge.ResourceTypeNotification,
			NativeID:     fmt.Sprintf("%d", it.ID),
		}

		results = append(results, notification.Notification{
			ID:           fmt.Sprintf("%d", it.ID),
			AccountID:    c.acc.ID,
			ResourceID:   resID,
			RepoOwner:    it.Repository.Owner.Username,
			RepoName:     it.Repository.Name,
			SubjectTitle: it.Subject.Title,
			SubjectType:  subType,
			Reason:       "mention",
			Unread:       it.Unread,
			UpdatedAt:    it.UpdatedAt,
			HTMLURL:      fmt.Sprintf("https://%s/%s/%s", c.acc.Host, it.Repository.Owner.Username, it.Repository.Name),
		})
	}

	return results, shared.Pagination{Page: 1, PerPage: len(results)}, nil
}

var _ ports.ForgeClient = (*Client)(nil)
