---
name: ddev-mcp
description: Manage DDEV projects via MCP tools. Use when the user needs DDEV commands or a DDEV project (.ddev/ directory) is present. Uses ddev-addon-mcp MCP tools instead of direct CLI calls.
---

# DDEV MCP Skill

Manages DDEV development environments via the MCP tools provided by ddev-addon-mcp. Prefer MCP tools over direct `ddev` CLI calls — they return structured responses and are optimized for agents.

## Prerequisites

The DDEV MCP add-on must be installed. Check if the MCP tools are available:

1. `.mcp.json` in the project must reference `ddev-mcp-server`
2. Or `.ddev/bin/ddev-mcp-server` must exist

If not installed:
```bash
ddev addon get ddev-addon-mcp
```

## Available MCP Tools

### Project Lifecycle

| Tool | When to use |
|------|-------------|
| `ddev_start` | Start a project, spin up containers |
| `ddev_stop` | Stop a project |
| `ddev_restart` | After config changes in `.ddev/` |
| `ddev_poweroff` | Stop ALL projects (requires confirm: true!) |

### Project Information

| Tool | When to use |
|------|-------------|
| `ddev_list` | Overview of all DDEV projects, check status |
| `ddev_describe` | Details: URLs, DB credentials, PHP version, services |

### Command Execution

| Tool | When to use |
|------|-------------|
| `ddev_exec` | Run any command in the container (PHP, Node, etc.) |
| `ddev_composer` | Composer commands (require, update, install) |

### Database

| Tool | When to use |
|------|-------------|
| `ddev_import_db` | Import SQL dump (requires confirm: true!) |
| `ddev_export_db` | Export database |
| `ddev_snapshot` | Create or restore snapshots |

### Logs

| Tool | When to use |
|------|-------------|
| `ddev_logs` | View logs from web, db, or other services |

### Tool-Packs (if enabled in `.ddev/mcp-config.yaml`)

**Redis:**
- `ddev_redis_cli` — Execute redis-cli commands
- `ddev_redis_info` — Redis server info
- `ddev_redis_flush` — Flush Redis (requires confirm: true)

**Solr:**
- `ddev_solr_status` — Core status
- `ddev_solr_reload` — Reload a core
- `ddev_solr_query` — Execute Solr queries

**Mailhog:**
- `ddev_mailhog_list` — List captured emails
- `ddev_mailhog_search` — Search emails
- `ddev_mailhog_delete` — Delete all emails (requires confirm: true)

---

## Common Workflows

### Start a project and check status

1. Call `ddev_start` with `project: "projectname"`
2. Call `ddev_describe` for URLs and DB credentials
3. Open URL in browser or share with the user

### TYPO3: Flush cache

```
ddev_exec({ command: "vendor/bin/typo3 cache:flush" })
```

### TYPO3: Install an extension

```
ddev_composer({ command: "require typo3/cms-dashboard" })
ddev_exec({ command: "vendor/bin/typo3 extension:setup" })
```

### Laravel: Run migrations

```
ddev_exec({ command: "php artisan migrate" })
```

### WordPress: List plugins

```
ddev_exec({ command: "wp plugin list" })
```

### Back up database before changes

```
ddev_snapshot({ name: "before-migration" })
```

### Restore database

```
ddev_snapshot({ action: "restore", name: "before-migration" })
```

### Import SQL dump

```
ddev_import_db({ file: "/path/to/dump.sql", confirm: true })
```

### Check logs for errors

```
ddev_logs({ service: "web", tail: 100 })
```

### List all projects

```
ddev_list()
```

---

## Safety Rules

- **Always create a `ddev_snapshot`** before destructive database operations
- **`confirm: true`** is required for: `ddev_poweroff`, `ddev_import_db`, `ddev_redis_flush`, `ddev_mailhog_delete`
- **`ddev_poweroff` stops ALL projects** — warn the user before executing
- For `ddev_import_db`, inform the user that the existing database will be overwritten

## Troubleshooting

### Project won't start
1. `ddev_logs({ service: "web" })` — Check web container logs
2. `ddev_describe()` — Check status and config
3. `ddev_restart()` — Try a restart

### Docker not available
If MCP tools respond with "Docker" errors:
- Ask the user to start Docker Desktop
- Then retry `ddev_start`

### Database issues
1. `ddev_snapshot({ name: "debug-backup" })` — Create a backup
2. `ddev_logs({ service: "db" })` — Check DB logs
3. If needed: `ddev_snapshot({ action: "restore", name: "last-good-version" })`
