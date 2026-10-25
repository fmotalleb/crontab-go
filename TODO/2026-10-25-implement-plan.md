# Implement project reliability plan

## Objective
Apply the prioritized fixes from `PLAN.md`, add focused regression coverage, verify the repository, and commit only task-related changes.

## Todo
- [x] Correct Docker completion and exit-status handling.
- [x] Bound task and web-event concurrency; protect event listener snapshots.
- [x] Require an explicit web authentication policy and constrain event dispatch.
- [x] Add coordinated shutdown and make event sources cancellation-aware.
- [x] Improve logfile rotation/reopen behavior, config validation, parser I/O, and HTTP/task resource handling.
- [x] Align config/schema/docs and Go CI toolchain; repair invalid test fixture.
- [x] Review changes and run build, vet, unit, race, and lint checks.
- [x] Commit the intended files.

## Notes
- Concurrency continues to limit individual tasks; hooks retain the config-documented per-task behavior.
- Docker and process-signal integration were not available; those runtime limitations remain documented in `PLAN.md`.
- An unrelated untracked `crontab (Copy).txt` was present and is intentionally excluded.

## Verification
- `GOCACHE=/tmp/crontab-go-audit-cache go build ./...` — passed.
- `GOCACHE=/tmp/crontab-go-audit-cache go vet ./...` — passed.
- `GOCACHE=/tmp/crontab-go-audit-cache go test ./...` — passed.
- `GOCACHE=/tmp/crontab-go-audit-cache go test -race ./...` — passed.
- `GOCACHE=/tmp/crontab-go-audit-cache GOLANGCI_LINT_CACHE=/tmp/crontab-go-lint-cache go tool golangci-lint run ./...` — passed (0 issues).
