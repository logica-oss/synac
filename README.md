# synac

Sync agent configs from canonical sources, ported from `sync-agent-config.sh`.

## Usage

```bash
go run ./cmd/synac --help
go run ./cmd/synac --dry-run
go run ./cmd/synac --log-format json
go run ./cmd/synac --project-wide-source agents --path-specific-source off
```

## Config

`.synac.yaml` (preferred, supports comments) or `.synac.json` in the
repository root. YAML is recommended for human editing; JSON is supported
for machine generation. Environment variables use the `SYNAC_` prefix.
Precedence: flag > env > config file > default.

```yaml
dry-run: false
log-format: "console" # or "json"
project-wide-source: "github" # github, agents, or off
path-specific-source: "github" # github, claude, or off
skills-source: "agents" # agents, claude, or off
```

Each category selects the copy source. Defaults match the shell behavior.
`off` skips the category.
