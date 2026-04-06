package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var ClientID = "Iv23liFOc0ORYwW0WP4s"

// ResolveClientID returns the active Client ID, or an error if it has not
// been set (e.g. a fork that hasn't updated the default above).
func ResolveClientID() (string, error) {
	if ClientID == "" || ClientID == "YOUR_GITHUB_CLIENT_ID" {
		return "", errors.New("GitHub Client ID not configured — see config/config.go")
	}
	return ClientID, nil
}

// Config is the persisted user state stored in the OS config directory.
type Config struct {
	GitHubToken   string `json:"github_token,omitempty"`
	GitHubLogin   string `json:"github_login,omitempty"`
	LLMProvider   string `json:"llm_provider,omitempty"`
	LLMAPIKey     string `json:"llm_api_key,omitempty"`
	LLMBaseURL    string `json:"llm_base_url,omitempty"`
	DocOutputType string `json:"doc_output_type,omitempty"`
	DocOutputPath string `json:"doc_output_path,omitempty"`
}

func (c *Config) IsRegistered() bool {
	return c.GitHubToken != "" && c.GitHubLogin != ""
}

func (c *Config) IsGitHubSetUp() bool  { return c.IsRegistered() }
func (c *Config) IsLLMSetUp() bool     { return c.LLMProvider != "" }
func (c *Config) IsDocTypeSetUp() bool { return c.DocOutputType != "" }
func (c *Config) IsDocPathSetUp() bool { return c.DocOutputPath != "" }

func (c *Config) IsFullySetUp() bool {
	return c.IsGitHubSetUp() && c.IsLLMSetUp() && c.IsDocTypeSetUp() && c.IsDocPathSetUp()
}

func (c *Config) Reset() error {
	c.GitHubToken = ""
	c.GitHubLogin = ""
	c.LLMProvider = ""
	c.LLMAPIKey = ""
	c.LLMBaseURL = ""
	c.DocOutputType = ""
	c.DocOutputPath = ""
	return c.Save()
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "opendocs", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{}, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
