package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	mu sync.RWMutex

	ServerURL    string `json:"server_url"`
	DashboardURL string `json:"dashboard_url"`
	DeviceKey    string `json:"device_key"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`

	path string
}

func configPath() string {
	return filepath.Join(os.Getenv("APPDATA"), "timepad", "config.json")
}

func Load() (*Config, error) {
	p := configPath()
	cfg := &Config{
		path:         p,
		ServerURL:    "http://localhost:8080/api/v1",
		DashboardURL: "http://localhost:5173",
	}

	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.path = p
	if cfg.ServerURL == "" {
		cfg.ServerURL = "http://localhost:8080/api/v1"
	}
	if cfg.DashboardURL == "" {
		cfg.DashboardURL = "http://localhost:5173"
	}
	return cfg, nil
}

func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o600)
}

func (c *Config) SetTokens(access, refresh string) {
	c.mu.Lock()
	c.AccessToken = access
	c.RefreshToken = refresh
	c.mu.Unlock()
	_ = c.Save()
}

func (c *Config) SetDeviceKey(key string) {
	c.mu.Lock()
	c.DeviceKey = key
	c.mu.Unlock()
	_ = c.Save()
}

func (c *Config) GetAccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.AccessToken
}

func (c *Config) GetRefreshToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.RefreshToken
}

func (c *Config) GetDeviceKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DeviceKey
}

func (c *Config) GetDashboardURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DashboardURL
}
