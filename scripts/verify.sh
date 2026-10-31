#!/bin/bash
set -euo pipefail

# Comparing inside the shell keeps a typo from silently skipping the check, which an "if: inputs.verify == 'true'" would allow.
case "$(printf '%s' "${VERIFY}" | tr '[:upper:]' '[:lower:]')" in false | no | 0)
  echo "changed=false" >> "${GITHUB_OUTPUT}"
  exit 0
  ;;
esac

# git diff misses untracked files, and synac creates them on a first run.
drift="$(git status --porcelain)"
if [[ -n "${drift}" ]]; then
  echo "::group::Agent configs are out of date" >&2
  git --no-pager diff --no-color >&2
  echo "::endgroup::" >&2
  echo "::error::Run synac locally and commit the result." >&2
  echo "changed=true" >> "${GITHUB_OUTPUT}"
  exit 1
fi

echo "changed=false" >> "${GITHUB_OUTPUT}"
