package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ClientID is the GitHub OAuth App client ID. Forks must replace this value;
// see ResolveClientID for the validation logic.
var ClientID = "Iv23liFOc0ORYwW0WP4s"

// ResolveClientID returns the active client ID or an error when it has not
// been configured (e.g. a fork that still holds the placeholder value).
func ResolveClientID() (string, error) {
	if ClientID == "" || ClientID == "YOUR_GITHUB_CLIENT_ID" {
		return "", errors.New("GitHub Client ID not configured — see config/config.go")
	}
	return ClientID, nil
}

// Config holds all persisted user state. It is stored as JSON in the
// OS-specific config directory under opendocs/config.json.
type Config struct {
	GitHubToken     string `json:"github_token,omitempty"`
	GitHubLogin     string `json:"github_login,omitempty"`
	LLMProvider     string `json:"llm_provider,omitempty"`
	LLMModel        string `json:"llm_model,omitempty"`
	LLMAPIKey       string `json:"llm_api_key,omitempty"`
	LLMBaseURL      string `json:"llm_base_url,omitempty"`
	DocOutputType   string `json:"doc_output_type,omitempty"`
	DocOutputPath   string `json:"doc_output_path,omitempty"`
	// SetupDismissed is set to true when the user explicitly clicks Start in
	// the setup wizard. Once set, subsequent launches skip the wizard and go
	// directly to the main menu regardless of how many steps were completed.
	// It is cleared by Reset() so the wizard reappears after a full reset.
	SetupDismissed bool `json:"setup_dismissed,omitempty"`
}

// IsRegistered reports whether a GitHub token and login are both present.
func (c *Config) IsRegistered() bool {
	return c.GitHubToken != "" && c.GitHubLogin != ""
}

// IsLLMSetUp reports whether an LLM provider and model have been selected.
func (c *Config) IsLLMSetUp() bool { return c.LLMProvider != "" && c.LLMModel != "" }

// IsDocTypeSetUp reports whether a documentation output type has been chosen.
func (c *Config) IsDocTypeSetUp() bool { return c.DocOutputType != "" }

// IsDocPathSetUp reports whether a documentation output path has been set.
func (c *Config) IsDocPathSetUp() bool { return c.DocOutputPath != "" }

// IsFullySetUp reports whether every required configuration field is present.
func (c *Config) IsFullySetUp() bool {
	return c.IsRegistered() && c.IsLLMSetUp() && c.IsDocTypeSetUp() && c.IsDocPathSetUp()
}

// Reset clears every field and persists the empty config to disk.
func (c *Config) Reset() error {
	*c = Config{}
	return c.Save()
}

// configPath returns the absolute path of the JSON config file.
func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "opendocs", "config.json"), nil
}

// Load reads the config from disk. If the file does not yet exist a zero-value
// Config is returned so callers can treat a missing file as first-run.
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

// Save writes the current config to disk as indented JSON. The config
// directory is created if it does not already exist (mode 0700).
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
