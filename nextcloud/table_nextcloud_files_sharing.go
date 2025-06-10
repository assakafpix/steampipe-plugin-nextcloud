// nextcloud_shares.go
package nextcloud

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
)

// ocsShare represents a single share object from the Files Sharing API
type ocsShare struct {
	ID                    string  `json:"id"`
	ShareType             int     `json:"share_type"`
	ShareWith             string  `json:"share_with"`
	ShareWithDisplayName  string  `json:"share_with_displayname"`
	Path                  string  `json:"path"`
	Permissions           int     `json:"permissions"`
	Password              *string `json:"password"`
	PublicUpload          bool    `json:"public_upload"`
	ExpireDate            *string `json:"expire_date"`
	URL                   string  `json:"url"`
	UIDOwner              string  `json:"uid_owner"`
	Owner                 string  `json:"displayname_owner"`
	TimeCreated           int     `json:"stime"`
	TimeModified          int     `json:"item_mtime"`
}

// ocsShareListResponse wraps the JSON envelope for the Shares API list
type ocsShareListResponse struct {
	Ocs struct {
		Meta struct {
			Status     string `json:"status"`
			StatusCode int    `json:"statuscode"`
			Message    string `json:"message"`
		} `json:"meta"`
		Data []ocsShare `json:"data"`
	} `json:"ocs"`
}

// tableNextcloudShare defines the schema and list/get configuration for share objects
func tableNextcloudShare() *plugin.Table {
	return &plugin.Table{
		Name:        "nextcloud_share",
		Description: "Nextcloud file shares (including public links)",
		List: &plugin.ListConfig{
			Hydrate: listShares,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getShare,
		},
		Columns: []*plugin.Column{
			{Name: "id", Type: proto.ColumnType_INT, Description: "Share ID", Transform: transform.FromField("ID")},
			{Name: "path", Type: proto.ColumnType_STRING, Description: "Path of the shared object", Transform: transform.FromField("Path")},
			{Name: "name_owner", Type: proto.ColumnType_STRING, Description: "Name of the owner", Transform: transform.FromField("Owner")},
			{Name: "password", Type: proto.ColumnType_STRING, Description: "Password protecting the share, if any", Transform: transform.FromField("Password")},
			{Name: "time_created", Type: proto.ColumnType_INT, Description: "Creation time of the share", Transform: transform.FromField("TimeCreated")},
			{Name: "time_modified", Type: proto.ColumnType_INT, Description: "Modified time of the share", Transform: transform.FromField("TimeModified")},
			{Name: "expire_date", Type: proto.ColumnType_STRING, Description: "Expiration date of the share, if set", Transform: transform.FromField("ExpireDate")},
			{Name: "share_with", Type: proto.ColumnType_STRING, Description: "UserID or groupID the resource is shared with", Transform: transform.FromField("ShareWith")},
			{Name: "share_with_displayname", Type: proto.ColumnType_STRING, Description: "User or group the resource is shared with", Transform: transform.FromField("ShareWithDisplayName")},
			{Name: "share_type", Type: proto.ColumnType_INT, Description: "Type of the share (0=user, 3=public link)", Transform: transform.FromField("ShareType")},
			{Name: "permissions", Type: proto.ColumnType_INT, Description: "Permission mask", Transform: transform.FromField("Permissions")},
			{Name: "public_upload", Type: proto.ColumnType_BOOL, Description: "Whether public upload is allowed", Transform: transform.FromField("PublicUpload")},
			{Name: "url", Type: proto.ColumnType_STRING, Description: "Public URL of the share", Transform: transform.FromField("URL")},
			{Name: "owner", Type: proto.ColumnType_STRING, Description: "Owner of the share", Transform: transform.FromField("UIDOwner")},
			
		},
	}
}

// listShares retrieves all shares from the Files Sharing API
func listShares(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := GetClient(ctx, d.Connection)
	if err != nil {
		return nil, err
	}
	endpoint := "ocs/v2.php/apps/files_sharing/api/v1/shares?format=json"
	resp, err := client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ocsShareListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding JSON Nextcloud Shares: %w", err)
	}
	if result.Ocs.Meta.Status != "ok" {
		return nil, fmt.Errorf("OCS API error: %s (code %d)", result.Ocs.Meta.Message, result.Ocs.Meta.StatusCode)
	}

	for _, share := range result.Ocs.Data {
		d.StreamListItem(ctx, share)
	}
	return nil, nil
}

// getShare retrieves a single share by ID
func getShare(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	qual := d.EqualsQuals["id"]
	if qual == nil {
		return nil, fmt.Errorf("id qualifier not provided")
	}
	id := qual.GetInt64Value()

	client, err := GetClient(ctx, d.Connection)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("ocs/v2.php/apps/files_sharing/api/v1/shares/%d?format=json", id)
	resp, err := client.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ocsShareListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding JSON Nextcloud Share detail: %w", err)
	}
	if result.Ocs.Meta.Status != "ok" || len(result.Ocs.Data) == 0 {
		return nil, fmt.Errorf("share with ID %d not found", id)
	}
	// API returns single-element array
	return result.Ocs.Data[0], nil
}
