# AGENTS.md

YAML-configured crontab replacement for Docker environments. Single Go module `github.com/fmotalleb/crontab-go`, Go 1.27, generic-heavy modern stdlib.

## Workflow (always)
- Always **review code after changing it**: re-read the diff, then run `go build`/`go vet`, `go test ./...`, and `go tool golangci-lint run ./...` before finishing.
- Always **commit changes after applying them** (stage only intended files; match the repo's conventional-commit style).

## Commands (all verified)
- `make ci` — full CI gate, matches `.github/workflows/build.yml`: `go mod tidy` → `go generate` → install goreleaser → `goreleaser build --snapshot` → `misspell -w` → `golangci-lint run --fix` → `go test -race` → `govulncheck` → `git diff` clean check.
- `make all` = `mod gen install build spell lint test`; `make precommit` = `all vuln`; `make ci` = `precommit diff`.
- `make run` = `go run .` — needs a config file: reads `./config.yaml` or `-c <file>` and panics with neither.
- `make test` — uses `-race` only when `CGO_ENABLED != 0`, emits `coverage.html`.
- Single test: `go test ./core/event/ -run TestName -v`.

## Gotchas
- `make lint` and `make spell` **auto-fix files** (`--fix`, `-w`); `make ci`/`make diff` then hard-fail on any uncommitted change. Commit fixer output before `make ci`.
- Never run bare `golangci-lint`/`misspell`/`govulncheck` — the global golangci-lint is v1 while `.golangci.yml` is v2 format. Use `go tool <name> ...` (versions pinned via go.mod `tool` directive).
- goimports `local-prefixes` in `.golangci.yml` is a stale copy-paste (`github.com/fmotalleb/junction`); group `github.com/fmotalleb/crontab-go` imports last, matching `.vscode/settings.json`.
- Formatting enforced by gofumpt + goimports; godot requires doc comments to end in a period; govet `enable-all`; funlen ≤100 lines.
- `make clean` wipes the module cache (`go clean -modcache`) — avoid casually.

## Architecture
- Entry: `main.go` → `cmd/root.go` (Cobra). Loads YAML via viper (+ godotenv `.env`) into `config.Config`, validates, starts `cron.New(cron.WithSeconds())`, then `jobs.InitializeJobs` wires tasks+hooks per job; optional echo webserver; blocks on `global.CTX().Done()`.
- Extensibility uses `generator.Core[I,O]` (`generator/generator.go`): a first-match registry whose `Get` selects an implementation. It selects **event emitters** (`core/event`: cron, interval, on-init, web-event, docker, logfile) and **tasks** (`core/task`: command, get, post). Emitters implement `abstraction.EventGenerator.BuildTickChannel(EventDispatcher)`.
- `core/global` = singleton signal-aware context; context keys in `ctxutils`. Build info injected by goreleaser ldflags into `github.com/fmotalleb/go-tools/git.*`.
- Generics are the norm (e.g. `global.Put/Get[T]`, `utils.List[T]`). No generated code and no `//go:generate` — `make gen` is a no-op.
- `dist/` and the root `crontab-go` binary are gitignored goreleaser artifacts; don't hand-edit.

## Config
- `config.doc.yaml` = fully commented reference; `config.example.yaml` = minimal example; `config.local.yaml` = uncommitted dev/test artifact (says "don't use as a reference"). `schema.json` is hand-maintained, not generated.
- Env vars (viper-mapped): `SHELL`, `SHELL_ARGS` (default `/bin/sh -c` on Linux, `cmd /c` on Windows), `LOG_*`, `WEBSERVER_*`/`LISTEN_*`.

## Testing
- stdlib `testing` + `github.com/alecthomas/assert/v2` (no testify, it's only indirect).
- No testdata/golden/integration suites. `core/cmd_connection` tests only assert connection-type selection — they never hit Docker.
- `core/common/retry_test.go` is fully commented out (dead file, ignore).
- Event-registry tests call `prepareState()` (puts a `cron.New()` into `global`); silence logs with `zap.NewNop()`; prefer `t.Context()`.
- `core/` subsystems `webserver`, `jobs`, `cmd_connection`, `docker`, `global`, `os_credential` have no tests today.
- README Codecov badge is stale — no CI coverage upload, coverage isn't gated.

## Release
- Pushing a `v*` tag triggers `goreleaser release` (needs `GITHUB_TOKEN`); builds multi-arch Docker images to `ghcr.io/fmotalleb/crontab-go` (slim + distroless). Requires a Docker daemon with `containerd-snapshotter` enabled.

## Session Logging
- Always keep a record of thoughts, todo list, and analysis of the job you are doing in `./TODO/<session>`.