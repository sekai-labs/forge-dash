package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	baseURL := "https://gitlab.com/api/v4"
	if acc.Host != "" && acc.Host != "gitlab.com" {
		baseURL = fmt.Sprintf("https://%s/api/v4", acc.Host)
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
	return forge.GitLabCapabilities()
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
		req.Header.Set("PRIVATE-TOKEN", c.token)
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
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return shared.User{}, err
	}

	return shared.User{
		Username:    dto.Username,
		DisplayName: dto.Name,
		AvatarURL:   dto.AvatarURL,
		HTMLURL:     dto.WebURL,
	}, nil
}

func (c *Client) ListRepositories(ctx context.Context, filter shared.QueryFilter) ([]repository.Repository, shared.Pagination, error) {
	path := "/projects?membership=true&per_page=50&order_by=updated_at"
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var projects []struct {
		ID                int64  `json:"id"`
		Name              string `json:"name"`
		PathWithNamespace string `json:"path_with_namespace"`
		Description       string `json:"description"`
		DefaultBranch     string `json:"default_branch"`
		WebURL            string `json:"web_url"`
		SSHURLToRepo      string `json:"ssh_url_to_repo"`
		HTTPURLToRepo     string `json:"http_url_to_repo"`
		Visibility        string `json:"visibility"`
		ForkedFromProject *any   `json:"forked_from_project"`
		Namespace         struct {
			Path string `json:"path"`
		} `json:"namespace"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []repository.Repository
	for _, p := range projects {
		parts := strings.Split(p.PathWithNamespace, "/")
		owner := p.Namespace.Path
		if len(parts) >= 2 {
			owner = strings.Join(parts[:len(parts)-1], "/")
		}

		results = append(results, repository.Repository{
			ID:            fmt.Sprintf("%d", p.ID),
			AccountID:     c.acc.ID,
			Provider:      account.ForgeGitLab,
			Host:          c.acc.Host,
			Owner:         owner,
			Name:          p.Name,
			FullName:      p.PathWithNamespace,
			Description:   p.Description,
			DefaultBranch: p.DefaultBranch,
			IsPrivate:     p.Visibility != "public",
			IsFork:        p.ForkedFromProject != nil,
			HTMLURL:       p.WebURL,
			SSHURL:        p.SSHURLToRepo,
			CloneURL:      p.HTTPURLToRepo,
		})
	}

	return results, shared.Pagination{Page: 1, PerPage: len(results)}, nil
}

func (c *Client) ListChanges(ctx context.Context, filter shared.QueryFilter) ([]change.ChangeRequest, shared.Pagination, error) {
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 30
	}

	v := url.Values{}
	v.Set("scope", "all")
	v.Set("per_page", strconv.Itoa(perPage))
	v.Set("order_by", "updated_at")
	if filter.State == "closed" {
		v.Set("state", "closed")
	} else if filter.State == "merged" {
		v.Set("state", "merged")
	} else {
		v.Set("state", "opened")
	}

	if filter.Query != "" {
		v.Set("search", filter.Query)
	}

	var pathBuilder strings.Builder
	pathBuilder.WriteString("/merge_requests?")
	pathBuilder.WriteString(v.Encode())
	path := pathBuilder.String()

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var mrs []struct {
		IID          int        `json:"iid"`
		ProjectID    int64      `json:"project_id"`
		Title        string     `json:"title"`
		Description  string     `json:"description"`
		State        string     `json:"state"`
		CreatedAt    time.Time  `json:"created_at"`
		UpdatedAt    time.Time  `json:"updated_at"`
		MergedAt     *time.Time `json:"merged_at"`
		ClosedAt     *time.Time `json:"closed_at"`
		WebURL       string     `json:"web_url"`
		SourceBranch string     `json:"source_branch"`
		TargetBranch string     `json:"target_branch"`
		Draft        bool       `json:"draft"`
		UserNotes    int        `json:"user_notes_count"`
		Author       struct {
			Username  string `json:"username"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
			WebURL    string `json:"web_url"`
		} `json:"author"`
		Labels []string `json:"labels"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mrs); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []change.ChangeRequest
	for _, mr := range mrs {
		owner, repo := parseWebURL(mr.WebURL)

		st := change.ChangeStateOpen
		if mr.State == "merged" {
			st = change.ChangeStateMerged
		} else if mr.State == "closed" {
			st = change.ChangeStateClosed
		} else if mr.Draft {
			st = change.ChangeStateDraft
		}

		var labels []shared.Label
		for _, l := range mr.Labels {
			labels = append(labels, shared.Label{Name: l})
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeGitLab,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypeMR,
			NativeID:     fmt.Sprintf("%d", mr.IID),
		}

		results = append(results, change.ChangeRequest{
			ResourceID:   resID,
			AccountID:    c.acc.ID,
			Number:       mr.IID,
			Title:        mr.Title,
			State:        st,
			RepoOwner:    owner,
			RepoName:     repo,
			SourceBranch: mr.SourceBranch,
			TargetBranch: mr.TargetBranch,
			Author: shared.User{
				Username:    mr.Author.Username,
				DisplayName: mr.Author.Name,
				AvatarURL:   mr.Author.AvatarURL,
				HTMLURL:     mr.Author.WebURL,
			},
			Body:         mr.Description,
			Labels:       labels,
			Comments:     mr.UserNotes,
			CreatedAt:    mr.CreatedAt,
			UpdatedAt:    mr.UpdatedAt,
			ClosedAt:     mr.ClosedAt,
			MergedAt:     mr.MergedAt,
			HTMLURL:      mr.WebURL,
			CIStatus:     change.CIStatusPassing,
			ReviewStatus: change.ReviewStatusNone,
		})
	}

	return results, shared.Pagination{Page: 1, PerPage: perPage, HasNextPage: len(results) == perPage}, nil
}

func (c *Client) GetChange(ctx context.Context, owner, repo string, number int) (*change.ChangeRequest, error) {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/merge_requests/%d", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var mr struct {
		IID          int        `json:"iid"`
		Title        string     `json:"title"`
		Description  string     `json:"description"`
		State        string     `json:"state"`
		CreatedAt    time.Time  `json:"created_at"`
		UpdatedAt    time.Time  `json:"updated_at"`
		MergedAt     *time.Time `json:"merged_at"`
		ClosedAt     *time.Time `json:"closed_at"`
		WebURL       string     `json:"web_url"`
		SourceBranch string     `json:"source_branch"`
		TargetBranch string     `json:"target_branch"`
		Draft        bool       `json:"draft"`
		ChangesCount string     `json:"changes_count"`
		Author       struct {
			Username  string `json:"username"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
			WebURL    string `json:"web_url"`
		} `json:"author"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, err
	}

	st := change.ChangeStateOpen
	if mr.State == "merged" {
		st = change.ChangeStateMerged
	} else if mr.State == "closed" {
		st = change.ChangeStateClosed
	} else if mr.Draft {
		st = change.ChangeStateDraft
	}

	return &change.ChangeRequest{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitLab,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypeMR,
			NativeID:     fmt.Sprintf("%d", mr.IID),
		},
		AccountID:    c.acc.ID,
		Number:       mr.IID,
		Title:        mr.Title,
		State:        st,
		RepoOwner:    owner,
		RepoName:     repo,
		SourceBranch: mr.SourceBranch,
		TargetBranch: mr.TargetBranch,
		Author: shared.User{
			Username:    mr.Author.Username,
			DisplayName: mr.Author.Name,
			AvatarURL:   mr.Author.AvatarURL,
			HTMLURL:     mr.Author.WebURL,
		},
		Body:      mr.Description,
		CreatedAt: mr.CreatedAt,
		UpdatedAt: mr.UpdatedAt,
		ClosedAt:  mr.ClosedAt,
		MergedAt:  mr.MergedAt,
		HTMLURL:   mr.WebURL,
	}, nil
}

func (c *Client) GetDiff(ctx context.Context, owner, repo string, number int) ([]shared.DiffFile, error) {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/merge_requests/%d/diffs", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var diffs []struct {
		OldPath     string `json:"old_path"`
		NewPath     string `json:"new_path"`
		NewFile     bool   `json:"new_file"`
		RenamedFile bool   `json:"renamed_file"`
		DeletedFile bool   `json:"deleted_file"`
		Diff        string `json:"diff"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&diffs); err != nil {
		return nil, err
	}

	var results []shared.DiffFile
	for _, d := range diffs {
		st := shared.FileStatusModified
		if d.NewFile {
			st = shared.FileStatusAdded
		} else if d.DeletedFile {
			st = shared.FileStatusDeleted
		} else if d.RenamedFile {
			st = shared.FileStatusRenamed
		}

		results = append(results, shared.DiffFile{
			OldPath: d.OldPath,
			NewPath: d.NewPath,
			Status:  st,
			Patch:   d.Diff,
		})
	}
	return results, nil
}

func (c *Client) ListCommits(ctx context.Context, owner, repo string, number int) ([]shared.Commit, error) {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/merge_requests/%d/commits", projectPath, number)
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
		ID           string    `json:"id"`
		ShortID      string    `json:"short_id"`
		Title        string    `json:"title"`
		AuthorName   string    `json:"author_name"`
		AuthorEmail  string    `json:"author_email"`
		AuthoredDate time.Time `json:"authored_date"`
		WebURL       string    `json:"web_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	var results []shared.Commit
	for _, cm := range commits {
		results = append(results, shared.Commit{
			SHA:      cm.ID,
			ShortSHA: cm.ShortID,
			Author: shared.User{
				Username:    cm.AuthorEmail,
				DisplayName: cm.AuthorName,
			},
			Message: cm.Title,
			Date:    cm.AuthoredDate,
			HTMLURL: cm.WebURL,
		})
	}
	return results, nil
}

func (c *Client) ListChecks(ctx context.Context, owner, repo string, ref string) ([]shared.CheckRun, error) {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	v := url.Values{}
	v.Set("ref", ref)
	v.Set("per_page", "1")
	var pathBuilder strings.Builder
	pathBuilder.WriteString("/projects/")
	pathBuilder.WriteString(projectPath)
	pathBuilder.WriteString("/pipelines?")
	pathBuilder.WriteString(v.Encode())
	path := pathBuilder.String()
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var pipelines []struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
		WebURL string `json:"web_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pipelines); err != nil || len(pipelines) == 0 {
		return nil, nil
	}

	pipelineID := pipelines[0].ID
	jobsPath := fmt.Sprintf("/projects/%s/pipelines/%d/jobs", projectPath, pipelineID)
	jReq, err := c.newRequest(ctx, http.MethodGet, jobsPath, nil)
	if err != nil {
		return nil, err
	}
	jResp, err := c.httpClient.Do(ctx, jReq)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(jResp.Body)

	var jobs []struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
		WebURL string `json:"web_url"`
	}
	if err := json.NewDecoder(jResp.Body).Decode(&jobs); err != nil {
		return nil, err
	}

	var results []shared.CheckRun
	for _, j := range jobs {
		var conc shared.CheckConclusion
		switch j.Status {
		case "success":
			conc = shared.CheckConclusionSuccess
		case "failed":
			conc = shared.CheckConclusionFailure
		case "canceled":
			conc = shared.CheckConclusionCancelled
		case "skipped":
			conc = shared.CheckConclusionSkipped
		default:
			conc = shared.CheckConclusionPending
		}

		st := shared.CheckStatusCompleted
		if j.Status == "running" {
			st = shared.CheckStatusInProgress
		} else if j.Status == "pending" {
			st = shared.CheckStatusQueued
		}

		results = append(results, shared.CheckRun{
			ID:         fmt.Sprintf("%d", j.ID),
			Name:       j.Name,
			Status:     st,
			Conclusion: conc,
			URL:        j.WebURL,
		})
	}
	return results, nil
}

func (c *Client) ListComments(ctx context.Context, owner, repo string, number int, isChange bool) ([]shared.Comment, error) {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	var path string
	if isChange {
		path = fmt.Sprintf("/projects/%s/merge_requests/%d/notes", projectPath, number)
	} else {
		path = fmt.Sprintf("/projects/%s/issues/%d/notes", projectPath, number)
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var notes []struct {
		ID        int64     `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		Author    struct {
			Username  string `json:"username"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
		} `json:"author"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&notes); err != nil {
		return nil, err
	}

	var results []shared.Comment
	for _, n := range notes {
		results = append(results, shared.Comment{
			ID: fmt.Sprintf("%d", n.ID),
			Author: shared.User{
				Username:    n.Author.Username,
				DisplayName: n.Author.Name,
				AvatarURL:   n.Author.AvatarURL,
			},
			Body:      n.Body,
			CreatedAt: n.CreatedAt,
		})
	}
	return results, nil
}

func (c *Client) CreateComment(ctx context.Context, owner, repo string, number int, body string, isChange bool) (*shared.Comment, error) {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	var path string
	if isChange {
		path = fmt.Sprintf("/projects/%s/merge_requests/%d/notes", projectPath, number)
	} else {
		path = fmt.Sprintf("/projects/%s/issues/%d/notes", projectPath, number)
	}

	req, err := c.newRequest(ctx, http.MethodPost, path, map[string]string{"body": body})
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var note struct {
		ID        int64     `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		Author    struct {
			Username  string `json:"username"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
		} `json:"author"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return nil, err
	}

	return &shared.Comment{
		ID: fmt.Sprintf("%d", note.ID),
		Author: shared.User{
			Username:    note.Author.Username,
			DisplayName: note.Author.Name,
			AvatarURL:   note.Author.AvatarURL,
		},
		Body:      note.Body,
		CreatedAt: note.CreatedAt,
	}, nil
}

func (c *Client) ListReviews(ctx context.Context, owner, repo string, number int) ([]review.Review, error) {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/merge_requests/%d/approvals", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var app struct {
		ApprovedBy []struct {
			User struct {
				Username  string `json:"username"`
				Name      string `json:"name"`
				AvatarURL string `json:"avatar_url"`
			} `json:"user"`
		} `json:"approved_by"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, err
	}

	var results []review.Review
	for i, a := range app.ApprovedBy {
		results = append(results, review.Review{
			ID: fmt.Sprintf("%d", i+1),
			Author: shared.User{
				Username:    a.User.Username,
				DisplayName: a.User.Name,
				AvatarURL:   a.User.AvatarURL,
			},
			State: review.ReviewStateApproved,
		})
	}
	return results, nil
}

func (c *Client) SubmitReview(ctx context.Context, owner, repo string, number int, state review.ReviewState, body string) error {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/merge_requests/%d/approve", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer httpclient.DrainAndClose(resp.Body)

	if body != "" {
		_, _ = c.CreateComment(ctx, owner, repo, number, body, true)
	}
	return nil
}

func (c *Client) MergeChange(ctx context.Context, owner, repo string, number int, method string, title, message string) error {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/merge_requests/%d/merge", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodPut, path, nil)
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
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/merge_requests/%d", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodPut, path, map[string]string{"state_event": "close"})
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
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/merge_requests/%d", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodPut, path, map[string]string{"state_event": "reopen"})
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

	v := url.Values{}
	v.Set("scope", "all")
	v.Set("per_page", strconv.Itoa(perPage))
	v.Set("order_by", "updated_at")
	if filter.State == "closed" {
		v.Set("state", "closed")
	} else {
		v.Set("state", "opened")
	}

	if filter.Query != "" {
		v.Set("search", filter.Query)
	}

	var pathBuilder strings.Builder
	pathBuilder.WriteString("/issues?")
	pathBuilder.WriteString(v.Encode())
	path := pathBuilder.String()

	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var issues []struct {
		IID         int        `json:"iid"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		State       string     `json:"state"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
		ClosedAt    *time.Time `json:"closed_at"`
		WebURL      string     `json:"web_url"`
		UserNotes   int        `json:"user_notes_count"`
		Author      struct {
			Username  string `json:"username"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
			WebURL    string `json:"web_url"`
		} `json:"author"`
		Labels []string `json:"labels"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []issue.Issue
	for _, iss := range issues {
		owner, repo := parseWebURL(iss.WebURL)

		st := issue.IssueStateOpen
		if iss.State == "closed" {
			st = issue.IssueStateClosed
		}

		var labels []shared.Label
		for _, l := range iss.Labels {
			labels = append(labels, shared.Label{Name: l})
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeGitLab,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypeIssue,
			NativeID:     fmt.Sprintf("%d", iss.IID),
		}

		results = append(results, issue.Issue{
			ResourceID: resID,
			AccountID:  c.acc.ID,
			Number:     iss.IID,
			Title:      iss.Title,
			State:      st,
			RepoOwner:  owner,
			RepoName:   repo,
			Author: shared.User{
				Username:    iss.Author.Username,
				DisplayName: iss.Author.Name,
				AvatarURL:   iss.Author.AvatarURL,
				HTMLURL:     iss.Author.WebURL,
			},
			Body:      iss.Description,
			Labels:    labels,
			Comments:  iss.UserNotes,
			CreatedAt: iss.CreatedAt,
			UpdatedAt: iss.UpdatedAt,
			ClosedAt:  iss.ClosedAt,
			HTMLURL:   iss.WebURL,
		})
	}

	return results, shared.Pagination{Page: 1, PerPage: perPage, HasNextPage: len(results) == perPage}, nil
}

func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*issue.Issue, error) {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/issues/%d", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var iss struct {
		IID         int        `json:"iid"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		State       string     `json:"state"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
		ClosedAt    *time.Time `json:"closed_at"`
		WebURL      string     `json:"web_url"`
		UserNotes   int        `json:"user_notes_count"`
		Author      struct {
			Username  string `json:"username"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
			WebURL    string `json:"web_url"`
		} `json:"author"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&iss); err != nil {
		return nil, err
	}

	st := issue.IssueStateOpen
	if iss.State == "closed" {
		st = issue.IssueStateClosed
	}

	return &issue.Issue{
		ResourceID: forge.ResourceID{
			Provider:     account.ForgeGitLab,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypeIssue,
			NativeID:     fmt.Sprintf("%d", iss.IID),
		},
		AccountID: c.acc.ID,
		Number:    iss.IID,
		Title:     iss.Title,
		State:     st,
		RepoOwner: owner,
		RepoName:  repo,
		Author: shared.User{
			Username:    iss.Author.Username,
			DisplayName: iss.Author.Name,
			AvatarURL:   iss.Author.AvatarURL,
			HTMLURL:     iss.Author.WebURL,
		},
		Body:      iss.Description,
		Comments:  iss.UserNotes,
		CreatedAt: iss.CreatedAt,
		UpdatedAt: iss.UpdatedAt,
		ClosedAt:  iss.ClosedAt,
		HTMLURL:   iss.WebURL,
	}, nil
}

func (c *Client) CloseIssue(ctx context.Context, owner, repo string, number int) error {
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/issues/%d", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodPut, path, map[string]string{"state_event": "close"})
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
	projectPath := url.QueryEscape(fmt.Sprintf("%s/%s", owner, repo))
	path := fmt.Sprintf("/projects/%s/issues/%d", projectPath, number)
	req, err := c.newRequest(ctx, http.MethodPut, path, map[string]string{"state_event": "reopen"})
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
	req, err := c.newRequest(ctx, http.MethodGet, "/todos?state=pending&per_page=50", nil)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	resp, err := c.httpClient.Do(ctx, req)
	if err != nil {
		return nil, shared.Pagination{}, err
	}
	defer httpclient.DrainAndClose(resp.Body)

	var todos []struct {
		ID     int64  `json:"id"`
		Action string `json:"action_name"`
		Target struct {
			Title string `json:"title"`
			IID   int    `json:"iid"`
		} `json:"target"`
		TargetType string    `json:"target_type"`
		CreatedAt  time.Time `json:"created_at"`
		TargetURL  string    `json:"target_url"`
		Project    struct {
			PathWithNamespace string `json:"path_with_namespace"`
		} `json:"project"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		return nil, shared.Pagination{}, err
	}

	var results []notification.Notification
	for _, td := range todos {
		parts := strings.Split(td.Project.PathWithNamespace, "/")
		owner, repo := "unknown", "unknown"
		if len(parts) >= 2 {
			owner = strings.Join(parts[:len(parts)-1], "/")
			repo = parts[len(parts)-1]
		}

		subType := notification.SubjectTypePR
		if td.TargetType == "Issue" {
			subType = notification.SubjectTypeIssue
		}

		resID := forge.ResourceID{
			Provider:     account.ForgeGitLab,
			Host:         c.acc.Host,
			RepoOwner:    owner,
			RepoName:     repo,
			ResourceType: forge.ResourceTypeNotification,
			NativeID:     fmt.Sprintf("%d", td.ID),
		}

		results = append(results, notification.Notification{
			ID:           fmt.Sprintf("%d", td.ID),
			AccountID:    c.acc.ID,
			ResourceID:   resID,
			RepoOwner:    owner,
			RepoName:     repo,
			SubjectTitle: td.Target.Title,
			SubjectType:  subType,
			Reason:       td.Action,
			Unread:       true,
			UpdatedAt:    td.CreatedAt,
			HTMLURL:      td.TargetURL,
		})
	}

	return results, shared.Pagination{Page: 1, PerPage: len(results)}, nil
}

func parseWebURL(webURL string) (string, string) {

	u, err := url.Parse(webURL)
	if err != nil {
		return "unknown", "unknown"
	}
	clean := strings.TrimPrefix(u.Path, "/")
	clean = strings.Split(clean, "/-/")[0]
	parts := strings.Split(clean, "/")
	if len(parts) >= 2 {
		return strings.Join(parts[:len(parts)-1], "/"), parts[len(parts)-1]
	}
	return "unknown", "unknown"
}

var _ ports.ForgeClient = (*Client)(nil)
