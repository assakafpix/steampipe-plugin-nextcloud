# Steampipe Plugin for Nextcloud

Query your Nextcloud instance with SQL using [Steampipe](https://steampipe.io).

## Developing

### Prerequisites
* [Steampipe](https://steampipe.io/downloads)
* [Golang](https://golang.org/doc/install)

### Clone

```bash
git clone git@github.com:assakafpix/steampipe-plugin-nextcloud.git
cd steampipe-plugin-nextcloud
```

### Build

Build, which automatically installs the new version to your `~/.steampipe/plugins` directory:

```bash
make
```

### Configure the plugin

```bash
cp config/* ~/.steampipe/config
nano ~/.steampipe/config/nextcloud.spc
```

### Try it!

```bash
steampipe query
> .inspect nextcloud
```

## Quick Start

Check activity:
```sql
select * from nextcloud_activity order by time desc;
```

