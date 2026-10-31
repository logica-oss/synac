# synac

![GitHub Release](https://img.shields.io/github/v/release/logica-oss/synac)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![autofix.ci](https://github.com/logica-oss/synac/actions/workflows/autofix.yaml/badge.svg)](https://github.com/logica-oss/synac/actions/workflows/autofix.yaml)
[![Verify (App)](https://github.com/logica-oss/synac/actions/workflows/verify-app.yaml/badge.svg)](https://github.com/logica-oss/synac/actions/workflows/verify-app.yaml)
[![Verify (Actions)](https://github.com/logica-oss/synac/actions/workflows/verify-actions.yaml/badge.svg)](https://github.com/logica-oss/synac/actions/workflows/verify-actions.yaml)
[![CodeQL Advanced](https://github.com/logica-oss/synac/actions/workflows/codeql.yaml/badge.svg)](https://github.com/logica-oss/synac/actions/workflows/codeql.yaml)
[![Trivy](https://github.com/logica-oss/synac/actions/workflows/trivy.yaml/badge.svg)](https://github.com/logica-oss/synac/actions/workflows/trivy.yaml)
[![codecov](https://codecov.io/gh/logica-oss/synac/graph/badge.svg)](https://codecov.io/gh/logica-oss/synac)

**Syn**c **a**gent **c**onfigs from canonical sources

## What is `synac`?

### What it does

- **Project-wide**: `.github/copilot-instructions.md` <-> `AGENTS.md`.
- **Path-specific**: `.github/instructions/` <-> `.claude/rules/`, converting frontmatter.
- **Skills**: `.agents/skills/` <-> `.claude/skills/`, copying `SKILL.md` trees.
- Stamps the copy with its source; `off` skips a category.

### What it is for

- Edit the source only.
- Enforce mirrors in CI, or auto-commit with `verify: "false"`.

## Install

```bash
go install github.com/logica-oss/synac/cmd/synac@latest
```

Or grab a binary from [Releases](https://github.com/logica-oss/synac/releases) (`linux`/`darwin` × `amd64`/`arm64`):

```bash
VERSION=vX.X.X # pick from Releases
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')

curl -fsSL "https://github.com/logica-oss/synac/releases/download/${VERSION}/synac_${VERSION#v}_${OS}_${ARCH}.tar.gz" | tar -xz synac
sudo install -m 755 synac /usr/local/bin/
```

## Usage

```plaintext
Usage: synac [options]

Sync agent configs from canonical sources.

Options:
  -c, --config string                 config file path (default: <root>/.synac.yaml or .synac.json)
  -n, --dry-run                       print planned changes without writing
      --log-format string             log format: console or json (default "console")
      --path-specific-source string   path-specific source: github, claude, or off (default "github")
      --project-wide-source string    project-wide source: github, agents, or off (default "github")
      --root string                   repository root (default: $GITHUB_WORKSPACE or git top level)
      --skills-source string          skills source: agents, claude, or off (default "agents")
  -v, --version                       print version and exit

Config file: .synac.yaml (preferred, supports comments) or .synac.json in the repository root.
Environment: SYNAC_* variables (e.g. SYNAC_DRY_RUN, SYNAC_LOG_FORMAT).
Precedence: flag > env > config file > default.
```

## Config

`.synac.yaml` (preferred) or `.synac.json` in the repo root.  
Precedence: flag > env (`SYNAC_` prefix) > file > default.

```yaml
dry-run: false
log-format: "console" # or "json"
project-wide-source: "github" # github, agents, or off
path-specific-source: "github" # github, claude, or off
skills-source: "agents" # agents, claude, or off
```

Each category selects the copy source (`off` skips it).

## GitHub Actions

`synac` can also be used as a composite action.  
By default, it fails if drift is detected.

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
      - uses: actions/checkout@latest
        with:
          persist-credentials: false
      - uses: logica-oss/synac@vX.X.X
```

Pin the action with a full-length commit SHA and select the tool version via the `version` input (defaults to `latest`, resolved via the GitHub Releases API).

Supports Linux/macOS on `x64`/`arm64`.

| Input     | Default  | Description                                                               |
| --------- | -------- | ------------------------------------------------------------------------- |
| `version` | `latest` | Synac version, for example `v0.1.0`, `v0`, or `latest`                    |
| `verify`  | `true`   | `false` syncs without failing, so a later step can commit the result      |
| `args`    | `""`     | Extra flags appended to the invocation, for example `--skills-source off` |

| Output    | Description                                              |
| --------- | -------------------------------------------------------- |
| `changed` | `true` when `verify` detected a drift, otherwise `false` |

Only `false`/`no`/`0` disables the check (fail-closed).

<details>
<summary>Auto-commit example</summary>

Needs `contents: write` with credentials kept.

```yaml
- uses: logica-oss/synac@vX.X.X
  with:
    verify: "false"
- run: |
    git config user.name "github-actions[bot]"
    git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
    git add -A
    git diff --cached --quiet || { git commit -m "chore: sync agent configs"; git push; }
```

</details>

## Pitfalls

Quote `*`-leading globs in frontmatter; bare `**/*.ts` is a YAML parse error.

```yaml
paths:
  - "**/*.ts" # good
  - **/*.ts # bad: parse error
```
