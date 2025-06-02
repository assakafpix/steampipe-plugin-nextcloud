package nextcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

// NextcloudConfig represents the connection configuration for Nextcloud.
// Only Basic Auth (username/password) is supported.
type NextcloudConfig struct {
	ServerURL *string `cty:"server_url"`
	Username  *string `cty:"username"`
	Password  *string `cty:"password"`
}

// NextcloudClient is an HTTP client for the Nextcloud OCS API.
type NextcloudClient struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

func ConfigInstance() interface{} {
	return &NextcloudConfig{}
}

// NewNextcloudClient creates and validates a new NextcloudClient.
// It now accepts *plugin.Connection so it can be called from Hydrate functions.
func NewNextcloudClient(ctx context.Context, conn *plugin.Connection) (*NextcloudClient, error) {
	// Load config from the plugin connection
	cfg := GetConfig(conn)

	client := &NextcloudClient{
		HTTPClient: &http.Client{
			Timeout: 30 *time.Second,
		},
	}

	// server_url must be provided
	if cfg.ServerURL != nil && *cfg.ServerURL != "" {
		client.BaseURL = *cfg.ServerURL
	} else {
		return nil, fmt.Errorf("server_url must be configured")
	}

	// username/password must be provided
	if cfg.Username != nil && *cfg.Username != "" {
		client.Username = *cfg.Username
	} else {
		return nil, fmt.Errorf("username must be configured")
	}
	if cfg.Password != nil && *cfg.Password != "" {
		client.Password = *cfg.Password
	} else {
		return nil, fmt.Errorf("password must be configured")
	}

	// Ensure BaseURL ends with "/"
	if !strings.HasSuffix(client.BaseURL, "/") {
		client.BaseURL += "/"
	}

	// Immediately test the connection
	if err := client.TestConnection(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to Nextcloud: %w", err)
	}

	return client, nil
}

// MakeRequest constructs and executes an HTTP request against Nextcloud’s OCS API.
func (c *NextcloudClient) MakeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	// Build full URL
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Required OCS headers
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Steampipe-Nextcloud-Plugin/1.0")

	// Basic Auth
	req.SetBasicAuth(c.Username, c.Password)

	// Execute
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Treat HTTP 4xx/5xx as error
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Nextcloud API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// GetJSON performs a GET request to the given endpoint and decodes the JSON response into 'result'.
func (c *NextcloudClient) GetJSON(ctx context.Context, endpoint string, result interface{}) error {
	resp, err := c.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

// TestConnection verifies the Nextcloud credentials by calling the capabilities endpoint.
func (c *NextcloudClient) TestConnection(ctx context.Context) error {
	var capabilities map[string]interface{}
	err := c.GetJSON(ctx, "ocs/v1.php/cloud/capabilities?format=json", &capabilities)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

func GetConfig(conn *plugin.Connection) *NextcloudConfig {
	if conn == nil || conn.Config == nil {
		return &NextcloudConfig{}
	}
	return conn.Config.(*NextcloudConfig)
}


// newConfigInstance returns a pointer to an empty NextcloudConfig.
func newConfigInstance() interface{} {
	return &NextcloudConfig{}
}

// GetClient builds and returns a validated NextcloudClient from *plugin.Connection.
func GetClient(ctx context.Context, conn *plugin.Connection) (*NextcloudClient, error) {
	return NewNextcloudClient(ctx, conn)
}


// Activity represents a single activity entry from Nextcloud’s Activity API.
type Activity struct {
	ID            int      `json:"id,string"`
	App           string   `json:"app"`
	Type          string   `json:"type"`
	Subject       string   `json:"subject"`
	SubjectRich   bool     `json:"subject_rich"`
	SubjectParams []string `json:"subject_params"`
	ObjectType    string   `json:"object_type"`
	ObjectID      string   `json:"object_id"`
	ObjectName    string   `json:"object_name"`
	Time          string   `json:"time"`
	Owner         string   `json:"owner"`
}
