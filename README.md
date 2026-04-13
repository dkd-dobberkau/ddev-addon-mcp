# ddev-addon-mcp

DDEV add-on that provides an MCP (Model Context Protocol) server for AI agent integration. Enables AI agents to manage your DDEV projects via stdio.

## Installation

```bash
ddev addon get ddev-addon-mcp
```

## Usage

### Claude Code

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "ddev": {
      "command": ".ddev/bin/ddev-mcp-server"
    }
  }
}
```

### Claude Desktop

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "ddev": {
      "command": "/path/to/project/.ddev/bin/ddev-mcp-server"
    }
  }
}
```

### Manual

```bash
ddev mcp serve    # Start MCP server (stdio)
ddev mcp status   # Show installation status
ddev mcp config   # Edit configuration
```

## Core Tools

| Tool | Description |
|------|-------------|
| `ddev_start` | Start a DDEV project |
| `ddev_stop` | Stop a DDEV project |
| `ddev_restart` | Restart a DDEV project |
| `ddev_poweroff` | Stop ALL projects (requires `confirm: true`) |
| `ddev_list` | List all projects with status |
| `ddev_describe` | Detailed project info |
| `ddev_exec` | Run commands in containers |
| `ddev_composer` | Run composer commands |
| `ddev_import_db` | Import SQL dump (requires `confirm: true`) |
| `ddev_export_db` | Export database |
| `ddev_snapshot` | Create/restore snapshots |
| `ddev_logs` | View service logs |

## Tool-Packs

Optional tools for additional services. Edit `.ddev/mcp-config.yaml`:

```yaml
toolpacks:
  redis: true    # redis-cli, info, flush
  solr: true     # status, reload, query
  mailhog: true  # list, search, delete emails
```

Tool-packs only load if the service is actually running.

## Requirements

- DDEV >= v1.24.0

## License

MIT
