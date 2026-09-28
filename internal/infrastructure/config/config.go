package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/forgedash/forgedash/internal/domain/account"
)

type SectionConfig struct {
	Title   string `yaml:"title"`
	Filters string `yaml:"filters"`
	Limit   int    `yaml:"limit,omitempty"`
}

type PreviewConfig struct {
	Open     bool   `yaml:"open"`
	Position string `yaml:"position"`
	Width    int    `yaml:"width,omitempty"`
}

type Config struct {
	DefaultView     string            `yaml:"default_view"`
	DefaultMode     string            `yaml:"default_mode"`
	ActiveAccount   string            `yaml:"active_account,omitempty"`
	Theme           string            `yaml:"theme"`
	VimMode         bool              `yaml:"vim_mode"`
	AutoRefresh     bool              `yaml:"auto_refresh"`
	RefreshInterval int               `yaml:"refresh_interval"`
	Preview         PreviewConfig     `yaml:"preview"`
	Accounts        []account.Account `yaml:"accounts"`
	PRSections      []SectionConfig   `yaml:"pr_sections"`
	IssueSections   []SectionConfig   `yaml:"issue_sections"`
}

func DefaultConfig() *Config {
	return &Config{
		DefaultView:     "prs",
		DefaultMode:     "unified",
		Theme:           "default",
		VimMode:         true,
		AutoRefresh:     true,
		RefreshInterval: 60,
		Preview: PreviewConfig{
			Open:     true,
			Position: "right",
			Width:    50,
		},
		Accounts: []account.Account{},
		PRSections: []SectionConfig{
			{Title: "Needs Review", Filters: "is:open review-requested:@me", Limit: 25},
			{Title: "My PRs", Filters: "is:open author:@me", Limit: 25},
			{Title: "Subscribed", Filters: "is:open", Limit: 25},
		},
		IssueSections: []SectionConfig{
			{Title: "Assigned", Filters: "is:open assignee:@me", Limit: 25},
			{Title: "Created", Filters: "is:open author:@me", Limit: 25},
			{Title: "Mentioned", Filters: "is:open mentions:@me", Limit: 25},
		},
	}
}

func DefaultConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ".forgedash.yaml"
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "forgedash", "config.yaml")
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config syntax: %w", err)
	}

	for _, a := range cfg.Accounts {
		if err := a.Validate(); err != nil {
			return nil, fmt.Errorf("account %s invalid: %w", a.Name, err)
		}
	}

	return cfg, nil
}

func Save(path string, cfg *Config) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Init(path string) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("configuration file already exists at %s", path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	template := `default_view: prs
default_mode: unified
theme: default
vim_mode: true
auto_refresh: true
refresh_interval: 60
preview:
  open: true
  position: right
  width: 50
accounts:
  - id: gh-personal
    name: GitHub Personal
    provider: github
    host: github.com
    auth_ref: github
pr_sections:
  - title: Needs Review
    filters: "is:open review-requested:@me"
    limit: 25
  - title: My PRs
    filters: "is:open author:@me"
    limit: 25
  - title: Subscribed
    filters: "is:open"
    limit: 25
issue_sections:
  - title: Assigned
    filters: "is:open assignee:@me"
    limit: 25
  - title: Created
    filters: "is:open author:@me"
    limit: 25
  - title: Mentioned
    filters: "is:open mentions:@me"
    limit: 25
`
	return os.WriteFile(path, []byte(template), 0644)
}
