package nextcloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

// tableNextcloudActivity defines the “nextcloud_activity” table schema.
func tableNextcloudActivity() *plugin.Table {
	return &plugin.Table{
		Name:        "nextcloud_activity",
		Description: "Nextcloud activity events (from the Activity app)",
		List: &plugin.ListConfig{
			Hydrate: listActivity,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Activity ID", Transform: transform.FromField("ID")},
			{Name: "app", Type: proto.ColumnType_STRING, Description: "Originating app", Transform: transform.FromField("App")},
			{Name: "type", Type: proto.ColumnType_STRING, Description: "Activity type", Transform: transform.FromField("Type")},
			{Name: "subject", Type: proto.ColumnType_STRING, Description: "Unformatted subject", Transform: transform.FromField("Subject")},
			{Name: "subject_rich", Type: proto.ColumnType_BOOL, Description: "Subject contains HTML", Transform: transform.FromField("SubjectRich")},
			{Name: "subject_params", Type: proto.ColumnType_JSON, Description: "Parameters for rich subject", Transform: transform.FromField("SubjectParams")},
			{Name: "object_type", Type: proto.ColumnType_STRING, Description: "Type of object acted upon", Transform: transform.FromField("ObjectType")},
			{Name: "object_id", Type: proto.ColumnType_STRING, Description: "ID of the object", Transform: transform.FromField("ObjectID")},
			{Name: "object_name", Type: proto.ColumnType_STRING, Description: "Name of the object", Transform: transform.FromField("ObjectName")},
			{Name: "time", Type: proto.ColumnType_TIMESTAMP, Description: "Timestamp of the activity", Transform: transform.FromField("Time")},
			{Name: "owner", Type: proto.ColumnType_STRING, Description: "User who performed the action", Transform: transform.FromField("Owner")},
		},
	}
}

// ocsResponse wraps the JSON envelope returned by Nextcloud’s OCS API.
type ocsResponse struct {
	Ocs struct {
		Meta struct {
			Status     string `json:"status"`
			StatusCode int    `json:"statuscode"`
			Message    string `json:"message"`
		} `json:"meta"`
		Data []Activity `json:"data"`
	} `json:"ocs"`
}

// listActivity calls the Nextcloud Activity OCS endpoint and streams rows.
func listActivity(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Build client using d.Connection (type *plugin.Connection)
	client, err := GetClient(ctx, d.Connection)
	if err != nil {
		return nil, err
	}

	// Endpoint to fetch activity logs
	endpoint := "ocs/v2.php/apps/activity/api/v2/activity?format=json"

	// Perform GET request
	resp, err := client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode OCS JSON envelope
	var result ocsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Nextcloud activity JSON: %w", err)
	}

	// Stream each activity row
	for _, activity := range result.Ocs.Data {
		d.StreamListItem(ctx, activity)
	}

	return nil, nil
}
