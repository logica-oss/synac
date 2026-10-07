# Repository Instructions

- Attach exactly one of the `patch`, `minor`, or `major` labels when creating or updating a PR.
- Verify the label is attached with `gh pr view --json labels` before finishing.

## Fix

- After any Go change: `task tidy`, `task format`, `task lint`.
- Each of those rewrites files, so finish this section before checking.

## Check

- After any Go change: `task lint`, `task build`, `task test`.
- After any Actions change: `actionlint`, `ghalint run`, `ghalint run-action`, `zizmor`.
