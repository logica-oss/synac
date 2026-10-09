# synac

[![codecov](https://codecov.io/gh/logica-oss/synac/graph/badge.svg)](https://codecov.io/gh/logica-oss/synac)

Sync agent configs from canonical sources, ported from `sync-agent-config.sh`.

## Usage

```bash
go run ./cmd/synac --help
go run ./cmd/synac --dry-run
go run ./cmd/synac --log-format json
go run ./cmd/synac --project-wide-source agents --path-specific-source off
```

## GitHub Actions

`synac` runs as a composite action. It downloads the release binary that
matches the tag it is referenced by, verifies it against the release
checksum, then syncs the checked-out repository. By default it fails the
job when the sync changes anything, so CI catches agent configs that
drifted from their canonical sources.

```yaml
name: Sync agent configs
on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  sync:
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
        with:
          persist-credentials: false
      - uses: logica-oss/synac@v0.1.0
```

Pin a released tag. The action resolves the binary from
`github.action_ref`, so `@main`, a branch, or a commit SHA has no matching
release asset and the job fails. Linux and macOS runners on `x64` and
`arm64` are supported; Windows runners are not.

| Input    | Default | Description                                                               |
| -------- | ------- | ------------------------------------------------------------------------- |
| `verify` | `true`  | `false` syncs without failing, so a later step can commit the result      |
| `args`   | `""`    | Extra flags appended to the invocation, for example `--skills-source off` |

| Output    | Description                                              |
| --------- | -------------------------------------------------------- |
| `changed` | `true` when `verify` detected a drift, otherwise `false` |

`verify` follows the naming used by `golangci/golangci-lint-action`, and
`args` matches `astral-sh/ruff-action`.

`synac` itself is run for real even under `verify`, because its `--dry-run`
lists the files it would touch rather than whether they differ: the output
is identical for a clean tree and a drifted one. The action therefore
compares the working tree with `git status --porcelain` afterwards, which
also covers newly created files. Any change fails the job, not only the
files `synac` manages.

Set `verify: "false"` to sync and commit the result yourself. Only
`false`, `no` and `0` disable the check, so a typo fails closed rather
than silently skipping the verification.

Committing needs write access, so that job drops `persist-credentials`
and grants `contents: write` instead of the read-only default above. A
checkout that keeps its credentials is the exception here, because the
push needs them.

```yaml
name: Sync agent configs
on:
  push:
    branches: [main]

permissions:
  contents: write

jobs:
  sync:
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: logica-oss/synac@v0.1.0
        with:
          verify: "false"
      - name: Commit the sync result
        run: |
          git config user.name  "github-actions[bot]"
          git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
          git add -A
          # Only an empty change set is skipped, so a real commit failure still fails the job.
          if git diff --cached --quiet; then
            echo "already in sync"
            exit 0
          fi
          git commit -m "chore: sync agent configs"
          git push
```

Everything except `verify` belongs in `.synac.yaml`. The action resolves
the repository root from `GITHUB_WORKSPACE` and discovers the config file
there, so there is no input for either. Both inputs resolve above
`.synac.yaml`, so `args` can override a setting for a single job without
editing the file.

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

## Pitfalls

Quote globs starting with `*` in frontmatter. Bare `**/*.ts` is not valid
YAML and fails the sync with a parse error.

```yaml
paths:
  - "**/*.ts" # good
  - **/*.ts # bad: parse error
```
