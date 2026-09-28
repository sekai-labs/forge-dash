package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Store struct {
	filePath string
	mu       sync.RWMutex
}

func NewStore(customPath string) (*Store, error) {
	if customPath != "" {
		return &Store{filePath: customPath}, nil
	}

	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	dir := filepath.Join(dataHome, "forgedash")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	return &Store{
		filePath: filepath.Join(dir, "credentials.json"),
	}, nil
}

func (s *Store) Get(authRef string) (string, error) {

	envKey := fmt.Sprintf("FORGEDASH_%s_TOKEN", strings.ToUpper(strings.ReplaceAll(authRef, "-", "_")))
	if val := os.Getenv(envKey); val != "" {
		return val, nil
	}

	switch strings.ToLower(authRef) {
	case "github", "gh":
		if val := os.Getenv("GITHUB_TOKEN"); val != "" {
			return val, nil
		}
		if val := os.Getenv("GH_TOKEN"); val != "" {
			return val, nil
		}
	case "gitlab", "gl":
		if val := os.Getenv("GITLAB_TOKEN"); val != "" {
			return val, nil
		}
		if val := os.Getenv("GL_TOKEN"); val != "" {
			return val, nil
		}
	case "codeberg", "forgejo", "gitea":
		if val := os.Getenv("CODEBERG_TOKEN"); val != "" {
			return val, nil
		}
		if val := os.Getenv("FORGEJO_TOKEN"); val != "" {
			return val, nil
		}
		if val := os.Getenv("GITEA_TOKEN"); val != "" {
			return val, nil
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var creds map[string]string
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", err
	}

	return creds[authRef], nil
}

func (s *Store) Set(authRef, secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	creds := make(map[string]string)
	data, err := os.ReadFile(s.filePath)
	if err == nil {
		_ = json.Unmarshal(data, &creds)
	}

	creds[authRef] = secret
	updated, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, updated, 0600)
}

func (s *Store) Delete(authRef string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	creds := make(map[string]string)
	if err := json.Unmarshal(data, &creds); err != nil {
		return err
	}

	delete(creds, authRef)
	updated, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, updated, 0600)
}

func (s *Store) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	creds := make(map[string]string)
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	var keys []string
	for k := range creds {
		keys = append(keys, k)
	}
	return keys, nil
}
