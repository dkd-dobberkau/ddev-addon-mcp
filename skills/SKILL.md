---
name: ddev-mcp
description: DDEV-Projekte über MCP Tools verwalten. Verwende wenn der User DDEV-Befehle braucht oder ein DDEV-Projekt (.ddev/ Verzeichnis) vorhanden ist. Nutzt die ddev-addon-mcp MCP Tools statt direkter CLI-Aufrufe.
---

# DDEV MCP Skill

Verwaltet DDEV-Entwicklungsumgebungen über die MCP Tools von ddev-addon-mcp. Bevorzuge MCP Tools gegenüber direkten `ddev` CLI-Aufrufen — sie liefern strukturierte Antworten und sind für Agenten optimiert.

## Voraussetzungen

Das DDEV MCP Add-on muss installiert sein. Prüfe ob die MCP Tools verfügbar sind:

1. `.mcp.json` im Projekt muss `ddev-mcp-server` referenzieren
2. Oder `.ddev/bin/ddev-mcp-server` muss existieren

Falls nicht installiert:
```bash
ddev addon get ddev-addon-mcp
```

## Verfügbare MCP Tools

### Projekt-Lifecycle

| Tool | Wann verwenden |
|------|----------------|
| `ddev_start` | Projekt starten, Container hochfahren |
| `ddev_stop` | Projekt stoppen |
| `ddev_restart` | Nach Config-Änderungen in `.ddev/` |
| `ddev_poweroff` | ALLE Projekte stoppen (confirm: true nötig!) |

### Projekt-Informationen

| Tool | Wann verwenden |
|------|----------------|
| `ddev_list` | Überblick aller DDEV-Projekte, Status prüfen |
| `ddev_describe` | Details: URLs, DB-Credentials, PHP-Version, Services |

### Befehle ausführen

| Tool | Wann verwenden |
|------|----------------|
| `ddev_exec` | Beliebige Befehle im Container (PHP, Node, etc.) |
| `ddev_composer` | Composer-Befehle (require, update, install) |

### Datenbank

| Tool | Wann verwenden |
|------|----------------|
| `ddev_import_db` | SQL-Dump importieren (confirm: true nötig!) |
| `ddev_export_db` | Datenbank exportieren |
| `ddev_snapshot` | Snapshot erstellen oder wiederherstellen |

### Logs

| Tool | Wann verwenden |
|------|----------------|
| `ddev_logs` | Logs von Web, DB oder anderen Services ansehen |

### Tool-Packs (falls aktiviert in `.ddev/mcp-config.yaml`)

**Redis:**
- `ddev_redis_cli` — Redis-Befehle ausführen
- `ddev_redis_info` — Redis Server-Info
- `ddev_redis_flush` — Redis leeren (confirm: true)

**Solr:**
- `ddev_solr_status` — Core-Status
- `ddev_solr_reload` — Core neu laden
- `ddev_solr_query` — Solr-Abfragen

**Mailhog:**
- `ddev_mailhog_list` — Eingefangene E-Mails auflisten
- `ddev_mailhog_search` — E-Mails durchsuchen
- `ddev_mailhog_delete` — Alle E-Mails löschen (confirm: true)

---

## Typische Workflows

### Projekt starten und Status prüfen

1. `ddev_start` mit `project: "projektname"` aufrufen
2. `ddev_describe` für URLs und DB-Credentials
3. URL im Browser öffnen oder dem User mitteilen

### TYPO3-Projekt: Cache leeren

```
ddev_exec({ command: "vendor/bin/typo3 cache:flush" })
```

### TYPO3-Projekt: Extension installieren

```
ddev_composer({ command: "require typo3/cms-dashboard" })
ddev_exec({ command: "vendor/bin/typo3 extension:setup" })
```

### Laravel-Projekt: Migration ausführen

```
ddev_exec({ command: "php artisan migrate" })
```

### WordPress: Plugin-Liste

```
ddev_exec({ command: "wp plugin list" })
```

### Datenbank sichern vor Änderungen

```
ddev_snapshot({ name: "vor-migration" })
```

### Datenbank wiederherstellen

```
ddev_snapshot({ action: "restore", name: "vor-migration" })
```

### SQL-Dump importieren

```
ddev_import_db({ file: "/path/to/dump.sql", confirm: true })
```

### Logs prüfen bei Fehlern

```
ddev_logs({ service: "web", tail: 100 })
```

### Alle Projekte auflisten

```
ddev_list()
```

---

## Sicherheitsregeln

- **Immer `ddev_snapshot` erstellen** bevor destruktive Datenbank-Operationen durchgeführt werden
- **`confirm: true`** ist erforderlich für: `ddev_poweroff`, `ddev_import_db`, `ddev_redis_flush`, `ddev_mailhog_delete`
- **`ddev_poweroff` stoppt ALLE Projekte** — den User warnen bevor es ausgeführt wird
- Bei `ddev_import_db` den User informieren, dass die bestehende Datenbank überschrieben wird

## Fehlerbehebung

### Projekt startet nicht
1. `ddev_logs({ service: "web" })` — Web-Container Logs prüfen
2. `ddev_describe()` — Status und Config prüfen
3. `ddev_restart()` — Neustart versuchen

### Docker nicht verfügbar
Wenn MCP Tools mit "Docker" Fehlern antworten:
- User auffordern Docker Desktop zu starten
- Dann `ddev_start` erneut versuchen

### Datenbank-Probleme
1. `ddev_snapshot({ name: "debug-backup" })` — Sicherung erstellen
2. `ddev_logs({ service: "db" })` — DB-Logs prüfen
3. Falls nötig: `ddev_snapshot({ action: "restore", name: "letzte-gute-version" })`
