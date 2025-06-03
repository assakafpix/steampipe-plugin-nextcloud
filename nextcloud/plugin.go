package nextcloud

import (
    "context"

    "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
    "github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
    "github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

func Plugin(ctx context.Context) *plugin.Plugin {
    p := &plugin.Plugin{
        Name: "steampipe-plugin-nextcloud",
        ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
            NewInstance: ConfigInstance,
            Schema:      configSchema,
        },
        DefaultTransform: transform.FromGo().NullIfZero(),
        TableMap: map[string]*plugin.Table{
            "nextcloud_activity": tableNextcloudActivity(),
        },
    }

    return p
}

var configSchema = map[string]*schema.Attribute{
    "server_url": {
        Type: schema.TypeString,
    },
    "username": {
        Type: schema.TypeString,
    },
    "password": {
        Type: schema.TypeString,
    },
}
