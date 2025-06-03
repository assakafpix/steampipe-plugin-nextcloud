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

// NextcloudConfig représente la configuration de connexion (Basic Auth).
type NextcloudConfig struct {
	ServerURL *string `cty:"server_url"`
	Username  *string `cty:"username"`
	Password  *string `cty:"password"`
}

// NextcloudClient est un client HTTP pour l’API OCS de Nextcloud.
type NextcloudClient struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

// ConfigInstance retourne une instance vide de configuration.
// Steampipe appellera cette fonction pour initialiser conn.Config.
func ConfigInstance() interface{} {
	return &NextcloudConfig{}
}

// NewNextcloudClient crée et valide un NextcloudClient.
// On y passe *plugin.Connection pour récupérer la config.
func NewNextcloudClient(ctx context.Context, conn *plugin.Connection) (*NextcloudClient, error) {
	// Récupérer la config (pointer ou valeur)
	cfg := GetConfig(conn)

	client := &NextcloudClient{
		HTTPClient: &http.Client{
			Timeout: 30 *time.Second,
		},
	}

	// Vérifier que server_url est renseigné
	if cfg.ServerURL != nil && *cfg.ServerURL != "" {
		client.BaseURL = *cfg.ServerURL
	} else {
		return nil, fmt.Errorf("server_url must be configured")
	}

	// Vérifier que username et password sont renseignés
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

	// S’assurer que BaseURL se termine par "/"
	if !strings.HasSuffix(client.BaseURL, "/") {
		client.BaseURL += "/"
	}

	// Tester immédiatement la connexion
	if err := client.TestConnection(ctx); err != nil {
		return nil, fmt.Errorf("unable to connect to Nextcloud: %w", err)
	}

	return client, nil
}

// MakeRequest construit et exécute une requête HTTP vers l’API OCS de Nextcloud.
func (c *NextcloudClient) MakeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	// Construire l’URL complète
	u, err := url.Parse(c.BaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Créer la requête HTTP
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// En-têtes OCS requis
	req.Header.Set("OCS-APIREQUEST", "true")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Steampipe-Nextcloud-Plugin/1.0")

	// Basic Auth
	req.SetBasicAuth(c.Username, c.Password)

	// Exécuter la requête
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Traiter les statuts HTTP 4xx/5xx comme des erreurs
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Nextcloud API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// GetJSON effectue un GET et décode la réponse JSON dans 'result'.
func (c *NextcloudClient) GetJSON(ctx context.Context, endpoint string, result interface{}) error {
	resp, err := c.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

// TestConnection vérifie les identifiants en appelant l’endpoint capabilities.
func (c *NextcloudClient) TestConnection(ctx context.Context) error {
	// Exemple : ocs/v1.php/cloud/capabilities?format=json
	var capabilities map[string]interface{}
	err := c.GetJSON(ctx, "ocs/v1.php/cloud/capabilities?format=json", &capabilities)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	return nil
}

// GetConfig récupère un *NextcloudConfig à partir de conn.Config,
// qu’il s’agisse d’un pointeur ou d’une valeur.
func GetConfig(conn *plugin.Connection) *NextcloudConfig {
	if conn == nil || conn.Config == nil {
		return &NextcloudConfig{}
	}
	switch cfg := conn.Config.(type) {
	case *NextcloudConfig:
		return cfg
	case NextcloudConfig:
		return &cfg
	default:
		// En cas d’autre type (imprévu), retourner une config vide pour éviter le panic
		return &NextcloudConfig{}
	}
}

// GetClient construit et retourne un NextcloudClient validé.
func GetClient(ctx context.Context, conn *plugin.Connection) (*NextcloudClient, error) {
	return NewNextcloudClient(ctx, conn)
}

// Activity représente une entrée d’activité depuis l’API Activity de Nextcloud.
// On déclare SubjectRich comme interface{} pour accepter un tableau ou un bool selon la version de Nextcloud.
type Activity struct {
	ID            int         `json:"id,string"`
	App           string      `json:"app"`
	Type          string      `json:"type"`
	Subject       string      `json:"subject"`
	SubjectRich   interface{} `json:"subject_rich"`
	SubjectParams []string    `json:"subject_params"`
	ObjectType    string      `json:"object_type"`
	ObjectID      int      `json:"object_id"`
	ObjectName    string      `json:"object_name"`
	Time          time.Time      `json:"time"`
	Owner         string      `json:"owner"`
}
