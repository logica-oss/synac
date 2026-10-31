#!/bin/bash
set -euo pipefail

# EXTRA_ARGS is left unquoted so it word-splits into separate arguments.
# shellcheck disable=SC2086
"${SYNAC}" ${EXTRA_ARGS:-}
