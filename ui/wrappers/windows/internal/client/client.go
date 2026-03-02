package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"timepad/windows/internal/config"
)

type EventInput struct {
	AppName     string    `json:"app_name"`
	WindowTitle string    `json:"window_title"`
	URL         string    `json:"url"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsIdle      bool      `json:"is_idle"`
}

type ingestPayload struct {
	DeviceKey string       `json:"device_key"`
	Events    []EventInput `json:"events"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Client struct {
	http *http.Client
	cfg  *config.Config
}

func New(cfg *config.Config) *Client {
	return &Client{
		http: &http.Client{Timeout: 15 * time.Second},
		cfg:  cfg,
	}
}

func (c *Client) PostEvents(events []EventInput) error {
	if c.cfg.GetDeviceKey() == "" {
		return fmt.Errorf("no device_key — register this device in the dashboard")
	}
	if c.cfg.GetAccessToken() == "" {
		return fmt.Errorf("not authenticated")
	}

	log.Printf("client: posting %d event(s) to %s/events", len(events), c.cfg.ServerURL)
	body, err := json.Marshal(ingestPayload{DeviceKey: c.cfg.GetDeviceKey(), Events: events})
	if err != nil {
		return err
	}

	status, err := c.doPost("/events", body)
	if err != nil {
		return err
	}
	log.Printf("client: POST /events -> HTTP %d", status)

	if status == http.StatusUnauthorized {
		log.Println("client: token expired, refreshing")
		if err := c.refreshToken(); err != nil {
			return fmt.Errorf("token refresh: %w", err)
		}
		status, err = c.doPost("/events", body)
		if err != nil {
			return err
		}
		log.Printf("client: retry POST /events -> HTTP %d", status)
	}

	if status >= 400 {
		return fmt.Errorf("POST /events HTTP %d", status)
	}
	return nil
}

func (c *Client) doPost(path string, body []byte) (int, error) {
	req, err := http.NewRequest(http.MethodPost, c.cfg.ServerURL+path, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.GetAccessToken())

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *Client) refreshToken() error {
	log.Println("client: calling POST /auth/refresh")
	rt := c.cfg.GetRefreshToken()
	if rt == "" {
		return fmt.Errorf("no refresh token")
	}
	body, _ := json.Marshal(map[string]string{"refresh_token": rt})
	req, err := http.NewRequest(http.MethodPost, c.cfg.ServerURL+"/auth/refresh", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server %d: %s", resp.StatusCode, data)
	}
	var r refreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	c.cfg.SetTokens(r.AccessToken, r.RefreshToken)
	log.Println("client: tokens refreshed and saved")
	return nil
}
