package forge

import (
	"fmt"
	"strings"

	"github.com/forgedash/forgedash/internal/domain/account"
)

type ResourceType string

const (
	ResourceTypePR           ResourceType = "pull"
	ResourceTypeMR           ResourceType = "merge_request"
	ResourceTypeIssue        ResourceType = "issue"
	ResourceTypeNotification ResourceType = "notification"
	ResourceTypeRepo         ResourceType = "repo"
)

type ResourceID struct {
	Provider     account.ForgeType `json:"provider"`
	Host         string            `json:"host"`
	RepoOwner    string            `json:"repo_owner"`
	RepoName     string            `json:"repo_name"`
	ResourceType ResourceType      `json:"resource_type"`
	NativeID     string            `json:"native_id"`
}

func (r ResourceID) URI() string {
	return "forgedash://" + string(r.Provider) + "/" + r.Host + "/" + r.RepoOwner + "/" + r.RepoName + "/" + string(r.ResourceType) + "/" + r.NativeID
}

func (r ResourceID) String() string {
	return r.URI()
}

func ParseResourceID(uri string) (ResourceID, error) {
	if !strings.HasPrefix(uri, "forgedash://") {
		return ResourceID{}, fmt.Errorf("invalid forgedash URI: %s", uri)
	}
	body := strings.TrimPrefix(uri, "forgedash://")
	parts := strings.Split(body, "/")
	if len(parts) < 6 {
		return ResourceID{}, fmt.Errorf("malformed forgedash URI, expected 6 segments: %s", uri)
	}

	return ResourceID{
		Provider:     account.ForgeType(parts[0]),
		Host:         parts[1],
		RepoOwner:    parts[2],
		RepoName:     parts[3],
		ResourceType: ResourceType(parts[4]),
		NativeID:     parts[5],
	}, nil
}
