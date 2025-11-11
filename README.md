# op-connect-secret-driver
#### _A Docker Secret driver for [1Password Connect][1PasswordConnectServer]_

This Docker Secret driver plugin integrates with **[1Password Connect server][1PasswordConnectServer]** to securely manage secrets in Docker Swarm.

## Requirements

* Docker Engine with Swarm mode enabled
* Docker Secret driver support
* [1Password Connect server][1PasswordConnectServer] setup and running
* 1Password Connect Token
* `1password-credentials.json` file in [data/1password-credentials.json](data/1password-credentials.json)

> **Note**: Unix socket creation is only supported on Linux and FreeBSD due to limitations in the "go-plugins-helpers" package.

## Configuration

### Connection to 1Password Connect

The SDK requires these environment variables to connect to [1Password Connect][1PasswordConnectServer]:

* `OP_CONNECT_HOST`: URL of your 1Password Connect server
* `OP_CONNECT_TOKEN`: Your 1Password Connect authentication token

Set them as Docker plugin configuration

```shell
docker plugin set op-connect-secret-driver:latest OP_CONNECT_HOST=http://localhost:17450 
docker plugin set op-connect-secret-driver:latest OP_CONNECT_TOKEN=your-1password-connect-token
```

### Docker Secret Driver Configuration

The plugin supports two ways to reference secrets:
1. Individual fields using `vault`, `item`, and optional `field` and `section` provided as secret labels
2. [1Password][1Password] URL format using the `ref` as secret label in the format `op://vault/item/field` or 
   `op://vault/item/section/field` (that you can copy from [1Password][1Password] directly)

Notes:
- The field parameter is optional and defaults to "password" if not specified
- The section parameter is optional and should be specified when the field is inside a section
- The plugin can retrieve both field values and file contents from [1Password][1Password] items
- When a section is specified, only fields within that section will be matched
- All configuration is done through labels

Example Docker Compose configurations:

```yaml
# Option 1: Using individual fields without section
secrets:
  db_password:
    driver: op-connect-secret-driver
    labels:
      vault: "your-vault-uuid-or-name"             # Required: Vault UUID or name
      item: "your-item-uuid-or-name"               # Required: Item UUID or name
      field: "password"                            # Optional: Defaults to "password"

# Option 2: Using individual fields with section
secrets:
  db_password:
    driver: op-connect-secret-driver
    labels:
      vault: "your-vault-uuid-or-name"             # Required: Vault UUID or name
      item: "your-item-uuid-or-name"               # Required: Item UUID or name
      section: "section-name"                      # Optional: Section name
      field: "password"                            # Optional: Defaults to "password"

# Option 3: Using 1Password URL reference without section
secrets:
  db_password:
    driver: op-connect-secret-driver
    labels:
      ref: "op://vault-name/item-name/field-name"  # Required: 1Password URL format

# Option 4: Using 1Password URL reference with section
secrets:
  db_password:
    driver: op-connect-secret-driver
    labels:
      ref: "op://vault-name/item-name/section-name/field-name"  # Required: 1Password URL format with section
```


## Installation from Docker Hub

The CI pipeline automatically builds and publishes the plugin to Docker Hub.
You can use this command to install the plugin:

### `linux/amd64`
```shell
docker plugin install clementmouchet/op-connect-secret-driver:linux-amd64 \
--grant-all-permissions \
--alias op-connect-secret-driver \
--disable
```

### `linux/arm64`
```shell
docker plugin install clementmouchet/op-connect-secret-driver:linux-arm64 \
--grant-all-permissions \
--alias op-connect-secret-driver \
--disable
```

## Build

You can also develop, build your own and install it locally.

### Recommended: Docker Build

```shell
docker compose build op-connect-secret-driver
docker compose up -d op-connect-secret-driver
docker compose cp op-connect-secret-driver:/op-connect-secret-driver plugin/rootfs/op-connect-secret-driver
docker compose stop op-connect-secret-driver && docker compose rm -f op-connect-secret-driver
```

### Alternative: Local Build

```shell
go build -o plugin/rootfs/op-connect-secret-driver
```

## Installation of local build

There's an [install.sh](install.sh) script for this.

```shell
./install.sh
```

### Manual Installation

1. Create the plugin:
```shell
docker plugin create op-connect-secret-driver plugin
```

2. Configure the plugin:
```shell
docker plugin set op-connect-secret-driver:latest OP_CONNECT_HOST=http://localhost:17450 
docker plugin set op-connect-secret-driver:latest OP_CONNECT_TOKEN=your-1password-connect-token
```

3. Start 1Password Connect services:
```shell
docker compose up op-connect-api
```

4. Enable the plugin:
```shell
docker plugin enable op-connect-secret-driver:latest
```

### Modifying Plugin

To modify plugin settings, first disable:

```shell
docker plugin disable op-connect-secret-driver:latest
```

To modify plugin code, first remove it, build it and start the installation process again.:

```shell
docker plugin remove op-connect-secret-driver:latest
```

## Troubleshooting

1. Verify plugin status:
```shell
docker plugin ls
```

2. Check plugin logs (syslog) or inspect it:
```shell
docker plugin inspect op-connect-secret-driver:latest
```

3. Verify configuration:
```shell
docker plugin inspect op-connect-secret-driver:latest -f "{{ .Settings.Env }}"
```

4. Ensure [1Password Connect server][1PasswordConnectServer] is accessible at the configured host

[1PasswordConnectServer]: https://developer.1password.com/docs/connect/get-started/
[1Password]: https://1password.com
