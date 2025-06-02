STEAMPIPE_INSTALL_DIR ?= ~/.steampipe
BUILD_TAGS = netgo
install:
	go build -o $(STEAMPIPE_INSTALL_DIR)/plugins/hub.steampipe.io/plugins/turbot/nextcloud@latest/steampipe-plugin-nextcloud.plugin -tags "${BUILD_TAGS}" *.go

# Exclude Parliament IAM permissions
dev:
	go build -o $(STEAMPIPE_INSTALL_DIR)/plugins/hub.steampipe.io/plugins/turbot/nextcloud@latest/steampipe-plugin-nextcloud.plugin -tags "dev ${BUILD_TAGS}" *.go