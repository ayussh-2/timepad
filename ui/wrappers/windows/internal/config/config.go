package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

const (
	defaultServerURL    = "http://localhost:8080/api/v1"
	defaultDashboardURL = "http://localhost:5173"
)

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.Trim(strings.TrimSpace(line[idx+1:]), `"`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

func Load() (*Config, error) {
	p := configPath()

	exeDir := func() string {
		ex, err := os.Executable()
		if err != nil {
			return "."
		}
		return filepath.Dir(ex)
	}()
	loadDotEnv(filepath.Join(exeDir, ".env"))

	cfg := &Config{
		path:         p,
		ServerURL:    defaultServerURL,
		DashboardURL: defaultDashboardURL,
	}

	data, err := os.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
		cfg.path = p
	}

	if v := os.Getenv("TIMEPAD_SERVER_URL"); v != "" {
		cfg.ServerURL = v
	}
	if v := os.Getenv("TIMEPAD_DASHBOARD_URL"); v != "" {
		cfg.DashboardURL = v
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = defaultServerURL
	}
	if cfg.DashboardURL == "" {
		cfg.DashboardURL = defaultDashboardURL
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

func (c *Config) GetServerURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ServerURL
}

func (c *Config) GetDashboardURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.DashboardURL
}

func (c *Config) SetServerURL(url string) {
	c.mu.Lock()
	c.ServerURL = url
	c.mu.Unlock()
	_ = c.Save()
}

func (c *Config) SetDashboardURL(url string) {
	c.mu.Lock()
	c.DashboardURL = url
	c.mu.Unlock()
	_ = c.Save()
}
