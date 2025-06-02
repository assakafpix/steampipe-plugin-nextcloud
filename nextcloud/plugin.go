package nextcloud

import (
    "context"

    "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
    "github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

// Plugin est le point d’entrée du plugin Nextcloud pour Steampipe.
func Plugin(ctx context.Context) *plugin.Plugin {
    return &plugin.Plugin{
        Name: "nextcloud",
        // On définit ici le schéma de connexion : 
        //   - NewInstance : fonction qui crée un *NextcloudConfig vierge
        //   - Schema : mappe chaque clé à son type et indique si elle est requise ou sensible
        ConnectionConfigSchema: &plugin.ConnectionConfigSchema{
            NewInstance: ConfigInstance,
            Schema: map[string]*schema.Attribute{
                "server_url": {
                    Type:     schema.TypeString,
                    Required: true,
                },
                "username": {
                    Type:     schema.TypeString,
                    Required: true,
                },
                "password": {
                    Type:      schema.TypeString,
                    Required:  true,
                },
            },
        },
        // La TableMap liste les tables exposées par ce plugin.
        TableMap: map[string]*plugin.Table{
            "nextcloud_activity": tableNextcloudActivity(),
        },
    }
}
