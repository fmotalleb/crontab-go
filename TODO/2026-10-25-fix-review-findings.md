# Fix review findings

## Objective
Verify the four supplied findings, preserve existing contracts except for the requested asynchronous web dispatch, and apply minimal fixes.

## Todo
- [x] Verify Docker cleanup, logfile startup/retry, web dispatch, and README findings against current code and callers.
- [x] Apply fixes and focused regression coverage.
- [x] Run build, vet, tests, race tests, lint, and review (limitations below).
- [x] Commit intended changes.

## Decisions
- All four findings are valid at baseline `1751eeab35fadc479f12e684ef7898d8f235a350`.
- Container removal gets a detached, ten-second context when deferred cleanup starts.
- Existing startup log content stays skipped; files opened on later retries start at byte zero.
- HTTP success confirms scheduling. The 64-dispatch limit remains held until every listener callback returns; each listener is scheduled independently.
- Keep asynchronous listener panic recovery so callback panics cannot terminate the server.

## Verification
- Docker cleanup regressions passed using a synthetic Docker API transport; no live daemon was required.
- Logfile startup/retry and independent, bounded web dispatch regressions passed.
- Build, vet, the full ordinary test suite, and the full race suite passed.
- Full pinned lint crashed in staticcheck's analysis of Go 1.27's standard library. The remaining configured linters passed with zero issues using `go tool golangci-lint run --disable staticcheck ./...`.
- `go tool govulncheck ./...` reported 14 vulnerabilities in the unchanged Go 1.27.0 standard library and Docker dependency; dependency upgrades are outside these four fixes.
- Reviewed the production diff and new tests; corrected one test assertion type mismatch found by compilation.
