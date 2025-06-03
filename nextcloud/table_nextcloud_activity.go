package nextcloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

// ocsActivityListResponse wrappe l’enveloppe JSON renvoyée par l’API Activity.
type ocsActivityListResponse struct {
	Ocs struct {
		Meta struct {
			Status     string `json:"status"`
			StatusCode int    `json:"statuscode"`
			Message    string `json:"message"`
		} `json:"meta"`
		Data []Activity `json:"data"`
	} `json:"ocs"`
}

// tableNextcloudActivity définit le schéma de la table "nextcloud_activity".
func tableNextcloudActivity() *plugin.Table {
	return &plugin.Table{
		Name:        "nextcloud_activity",
		Description: "Nextcloud activity events (from the Activity app)",
		List: &plugin.ListConfig{
			Hydrate: listActivity,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getActivity,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Activity ID", Transform: transform.FromField("ID")},
			{Name: "app", Type: proto.ColumnType_STRING, Description: "Originating app", Transform: transform.FromField("App")},
			{Name: "type", Type: proto.ColumnType_STRING, Description: "Activity type", Transform: transform.FromField("Type")},
			{Name: "subject", Type: proto.ColumnType_STRING, Description: "Unformatted subject", Transform: transform.FromField("Subject")},
			{Name: "subject_rich", Type: proto.ColumnType_JSON, Description: "Subject contains HTML (raw JSON)", Transform: transform.FromField("SubjectRich")},
			{Name: "subject_params", Type: proto.ColumnType_JSON, Description: "Parameters for rich subject", Transform: transform.FromField("SubjectParams")},
			{Name: "object_type", Type: proto.ColumnType_STRING, Description: "Type of object acted upon", Transform: transform.FromField("ObjectType")},
			{Name: "object_id", Type: proto.ColumnType_INT, Description: "ID of the object", Transform: transform.FromField("ObjectID")},
			{Name: "object_name", Type: proto.ColumnType_STRING, Description: "Name of the object", Transform: transform.FromField("ObjectName")},
			{Name: "time", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp of the activity", Transform: transform.FromField("Time")},
			{Name: "owner", Type: proto.ColumnType_STRING, Description: "User who performed the action", Transform: transform.FromField("Owner")},
		},
	}
}

// listActivity appelle l’endpoint OCS pour lister toutes les activités.
func listActivity(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Construire le client à partir de d.Connection
	client, err := GetClient(ctx, d.Connection)
	if err != nil {
		return nil, err
	}

	// Endpoint Nextcloud Activity (format JSON)
	endpoint := "ocs/v2.php/apps/activity/api/v2/activity?format=json"

	// Appel HTTP GET
	resp, err := client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Décodage de l’enveloppe JSON
	var result ocsActivityListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("échec du décodage JSON Nextcloud Activity : %w", err)
	}

	// Vérification du statut OCS
	if result.Ocs.Meta.Status != "ok" {
		return nil, fmt.Errorf("erreur OCS API : %s (code : %d)", result.Ocs.Meta.Message, result.Ocs.Meta.StatusCode)
	}

	// Si un filtre "user_id = X" est présent, on ne diffuse que les activités correspondant à owner == userID
	if qual := d.EqualsQuals["user_id"]; qual != nil {
		userID := qual.GetStringValue()
		for _, activity := range result.Ocs.Data {
			if activity.Owner == userID {
				d.StreamListItem(ctx, activity)
			}
		}
	} else {
		// pas de filtre, on diffuse toutes les activités
		for _, activity := range result.Ocs.Data {
			d.StreamListItem(ctx, activity)
		}
	}

	return nil, nil
}

// getActivity récupère une activité précise via son ID.
func getActivity(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Extraction du qualifier "id" depuis d.EqualsQuals
	qual := d.EqualsQuals["id"]
	if qual == nil {
		return nil, fmt.Errorf("id qualifier not provided")
	}
	id := qual.GetInt64Value()

	// Construire le client Nextcloud
	client, err := GetClient(ctx, d.Connection)
	if err != nil {
		return nil, err
	}

	// Récupérer toutes les activités (filtrage côté client)
	endpoint := "ocs/v2.php/apps/activity/api/v2/activity?format=json"
	resp, err := client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Décodage de l’enveloppe JSON
	var result ocsActivityListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("échec du décodage JSON Nextcloud Activity : %w", err)
	}
	if result.Ocs.Meta.Status != "ok" {
		return nil, fmt.Errorf("OCS API error: %s (code: %d)", result.Ocs.Meta.Message, result.Ocs.Meta.StatusCode)
	}

	// Recherche de l’activité dont l’ID correspond
	for _, activity := range result.Ocs.Data {
		if int64(activity.ID) == id {
			return activity, nil
		}
	}

	// Si aucune activité trouvée
	return nil, fmt.Errorf("activity with ID %d not found", id)
}
