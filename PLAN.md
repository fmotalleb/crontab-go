# Reliability and security improvement plan

Audit date: 2026-10-25. Scope: startup and shutdown, event sources, job execution, command and HTTP tasks, configuration, web dispatch, parser, and CI. This is a source review; no runtime behavior was changed. Items below are ordered by impact. A regression test should be added with each fix.

## P0 — incorrect outcomes and uncontrolled execution

1. **Report Docker command failures correctly.** `core/cmd_connection/docker_attach.go:133-159` copies exec output but never calls `ContainerExecInspect` to obtain the exit code. `core/cmd_connection/docker_create.go:158-220` reads logs with `Follow: false`, never waits for termination, and never checks container exit status. Both can report success for a failed command; the create path can also return before the command finishes. Wait for completion, inspect exit codes, return an error on nonzero status, and retain streamed output. Remove the `ContainerStats` polling loop and close any stats response body if stats remain useful. Cover zero/nonzero exit, delayed output, timeout, and cleanup using a Docker API test server or a Docker integration test.

2. **Bound work before spawning it.** `core/jobs/task_handler.go:30-35` starts one goroutine per task per event; the pool is acquired inside each goroutine at `:60`. A burst of web, Docker, or log events can create an unbounded queue of blocked goroutines and memory use. Put a bounded queue or context-aware permit acquisition ahead of goroutine creation; define and document overflow behavior (block, reject, or drop with a metric). Test bursts larger than `concurrency` and cancellation while queued. Decide whether `concurrency` limits tasks or whole job runs; current job hooks execute once per *task* (`:64-75`), even though they are named job hooks. Implement and test the chosen semantics.

3. **Make web event dispatch safe and controlled.** `core/global/global_context.go:88-100` returns a mutable listener map after releasing its lock, while event registration writes to that map. `core/webserver/endpoint/event_dispatch.go:27-35` reads it and launches an unbounded goroutine per listener/request. Use a locked lookup that returns a copied listener slice, then apply the same bounded dispatch policy as jobs. Add a race test with simultaneous registration and HTTP dispatch, and a burst test. The endpoint currently returns 200 when work is merely queued; document or change that response contract.

4. **Require an explicit safe web exposure policy.** `core/webserver/webserver.go:51-63,93-97` serves a task-triggering endpoint with no authentication whenever either credential is absent; `engine.Any` allows all methods. A misconfigured externally bound address can expose command execution. Reject partial credentials, require explicit opt-in for unauthenticated event dispatch, restrict to POST, add request limits/rate limiting as appropriate, and document binding to a trusted interface or TLS-enabled proxy. Test anonymous, partial credential, and authenticated requests. Consider separate auth policy for `/metrics`.

## P1 — lost events and unreliable lifecycle

5. **Handle container shutdown and in-flight jobs.** `core/global/global_context.go:44` listens for `os.Kill` (which cannot be caught) and omits SIGTERM, the normal container stop signal. `cmd/root.go:48-83` starts cron and the web server but returns on context cancellation without stopping cron, shutting down HTTP, or waiting for running tasks; `webserver.Serve` uses a blocking `Start` without a shutdown path. Handle SIGTERM, stop event producers, gracefully shut down HTTP, drain or cancel queued/running work with a deadline, and then flush telemetry. Test SIGTERM and a slow running task.

6. **Make event sources cancellation aware.** `core/event/cron.go:62-102` leaves a scheduled cron job sending to an unbuffered channel after its receiver exits; a later tick can block forever. Remove the cron entry and select on context cancellation while sending. `core/event/interval.go:60-82` never stops its ticker. `core/event/docker.go` sleeps during reconnect without respecting cancellation. Test cancellation while a tick/send or reconnect is pending, and confirm no goroutine remains.

7. **Make log-file events survive ordinary file changes.** `core/event/logfile.go:95-136` exits permanently if the file is absent at startup and keeps the original open file across rename/rotation or truncation. Reopen with a context-aware retry, track file identity and offset, handle rotation/truncation, and define whether pre-existing lines are skipped. Test absent-then-created, appended lines, rotation, truncation, and shutdown. Reject negative `log-check-cycle` before `time.NewTicker` can panic (`config/job_validator.go:128-170`, `core/event/logfile.go:63-65`).

8. **Treat malformed Docker event streams as reconnect conditions.** `core/event/docker.go` receives from `msg` and `errs` without checking channel closure. A closed channel can produce a busy loop of zero-value messages/errors. Check `ok`, reconnect or stop according to policy, and use cancellable delays. Test closed channels and repeated connection failures.

## P2 — configuration, parser, and observability quality

9. **Validate configuration before applying it.** `config/validators.go:20-25` dereferences every job without rejecting nil, and `config/job_validator.go:73-82` accepts an enabled job with no tasks or events. Validate nonempty names, duplicate names, required tasks/events, nonnegative `debounce`, `retry-timeout`, `retry-max-delay`, `retry-jitter`, and `log-check-cycle`, plus valid connection mode and URL scheme. `cmd/root.go:145-157` currently validates before `defaulter.ApplyDefaults`; apply defaults first if validation should assess effective values. Replace panics in startup with contextual errors. Add table-driven invalid-config tests and a startup smoke test.

10. **Align configuration representations and CI toolchains.** `go.mod` declares Go 1.27.0 while `.github/workflows/build.yml` and `.github/workflows/release.yml` install Go 1.26, so CI depends on automatic toolchain downloads or fails when they are disabled. The pinned golangci-lint v2.12.2 also fails to decode Go 1.27 standard-library export data in this audit. Choose one supported Go version, use a compatible linter, and align CI, AGENTS.md, and release builds. `config/config.go` serializes Docker `ErrorLimit` as `error-limit`, while its mapstructure key and `schema.json` use `error-limit-count`; `Task.WorkingDirectory` serializes as `working-directory`, while loading and schema use `working-dir`. Fix tags, add a YAML round-trip test, and validate `schema.json` against the documented and generated examples. The schema currently omits web credentials and includes a root `variables` property absent from `Config`; reconcile those too.

11. **Preserve crontab import semantics.** `cmd/parser/cron_string.go:68-99` stores the current mutable `envTable` in each imported task, so later environment assignments can retroactively change earlier tasks. Clone it for each spec. `cmd/parser/parser.go:67-84` reads a file in one `Read`, which may return fewer bytes than requested, and does not close the file or output file. Use `io.ReadAll`/`os.ReadFile`, close output and propagate close errors. Add fixtures for changing env values, partial readers where practical, multiple schedules, and round-trip load of emitted YAML.

12. **Remove shared mutable execution state and reduce sensitive output.** `core/task/get.go` and `core/task/post.go` call `SetCancel` on reusable task objects for every execution; concurrent executions can overwrite that field. Pass cancellation per invocation or remove the unused `Cancelable` state. `core/task/get.go:48-58` and `post.go:53-63` put full URLs into traces, and command tasks record the command text; these may contain secrets. Define redaction and size limits for telemetry and request/command output. Add a race test for concurrent reuse and redaction tests.

13. **Improve HTTP task resource management.** `core/task/http.go:17-23` creates a fresh transport for every attempt, preventing connection reuse; no transport is closed. Use a shared client/transport per TLS policy with idle connection cleanup on shutdown. Decide on a default timeout when task `timeout` is omitted, and cap or stream response output according to documented policy. Test repeated requests, cancellation, and large responses.

14. **Repair the invalid-data test fixture.** `config/task_validator_test.go:69-79` expects JSON marshaling an *empty* `map[any]any` to fail, but there are no keys to reject and the test currently fails. Use a nonempty map with an unsupported key type and assert the actual validation error. Keep the HTTP tests runnable in an environment that allows loopback listeners; the audit sandbox rejects `httptest.NewServer` before those assertions can run.

## Delivery sequence and gates

1. Land P0 fixes in small commits: Docker exit status, bounded execution/job-hook semantics, then web listener safety and exposure policy. Add focused regression tests for each.
2. Land lifecycle and event-source fixes. Run `go test -race ./...` and shutdown/integration tests on Linux with Docker where available.
3. Land validation, schema/parser round trips, and HTTP/telemetry hardening. Regenerate or manually reconcile examples and README.
4. For each change, review the diff and run `go build ./...`, `go vet ./...`, `go test ./...`, and `go tool golangci-lint run ./...`. After toolchain alignment, run the full `make ci` gate on all supported CI platforms.

## Audit limitations

Docker behavior and signal handling were assessed from source, without a live Docker daemon or container shutdown test. The plan calls for those reproductions before considering the relevant fixes complete. The checks run for this audit and any environment limitations are recorded in `TODO/2026-10-25-project-audit.md`.

## Implementation status — 2026-10-25

Implemented the planned code, config, documentation, and CI changes in this change set. Concurrency continues to limit individual tasks; hooks retain their documented per-task behavior. Web event dispatch blocks until listener callbacks return, is limited to 64 concurrent requests, and requires Basic Auth unless the explicit unauthenticated opt-in is set. Shutdown now stops cron and drains tracked tasks with a deadline. Docker exit status helpers and job/web/config regression tests were added. The audit limitations still apply: Docker behavior was not exercised against a live daemon, and signal handling was not tested through a spawned process.

Validation completed: `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`, and `go tool golangci-lint run ./...`.
