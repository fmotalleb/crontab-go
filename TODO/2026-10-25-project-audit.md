# Project audit session

## Objective
Inspect the scheduler, event sources, task execution, config validation, web server, and build checks. Record evidence-backed risks and write a prioritized PLAN.md. Commit intended documentation changes.

## Todo
- [x] Trace startup, scheduling, and shutdown behavior.
- [x] Trace task and connection execution and failure handling.
- [x] Inspect configuration validation, web endpoints, and parser.
- [x] Run required build, vet, tests, and lint.
- [x] Write and review PLAN.md; commit documentation.

## Analysis notes
- Initial source inventory and clean git status confirmed.

## Findings and decisions
- Docker attach/create execution paths omit exit-status checks; create also reads logs without following or waiting.
- Job handler spawns tasks before taking a concurrency slot; hooks currently fire for each task.
- Web event listener lookup exposes a mutable map after its lock is released; web dispatch also has an open-auth default and no bounded request fan-out.
- SIGTERM is not registered and shutdown does not drain HTTP, cron, or running jobs.
- Log watcher lacks file creation/rotation recovery; event sources have several cancellation gaps.
- Config validation accepts empty enabled jobs and misses some negative durations; serialization/schema keys drift.
- Parser reuses a mutable environment map and performs a single file Read.
- PLAN.md groups fixes by impact and includes testable completion criteria. No runtime code was changed.

## Verification so far
- The default Go cache is read only in this sandbox. Reran with `GOCACHE=/tmp/crontab-go-audit-cache`.
- `go build ./...` and `go vet ./...` passed.
- `go test ./...` failed: `TestTaskValidate_InvalidPostData` uses an empty map that JSON marshals successfully; `core/task` HTTP tests cannot open a loopback listener in this sandbox (`operation not permitted`).
- `go tool golangci-lint run ./...` failed with `cannot decode "internal/goarch", export data version 5 is greater than maximum supported version 4`. This remained after setting `GOLANGCI_LINT_CACHE=/tmp/crontab-go-lint-cache`; the initial run also encountered a read-only default lint cache.
