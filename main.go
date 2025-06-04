package main

import (
    "github.com/assakafpix/steampipe-plugin-nextcloud/nextcloud"
    "github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        PluginFunc: nextcloud.Plugin,
    })
}
