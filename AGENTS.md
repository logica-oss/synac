<!-- DO NOT EDIT: Generated mirror of /.github/copilot-instructions.md. Edit /.github/copilot-instructions.md instead. -->

# Repository Instructions

- Attach exactly one of the `patch`, `minor`, or `major` labels when creating or updating a PR.
- Verify the label is attached with `gh pr view --json labels` before finishing.

## Verification

- After any Go change: `task tidy`, `task format`, `task lint`, `task build`, `task test`.
- After any Actions change: `actionlint`, `ghalint run`, `ghalint run-action`, `zizmor`.
