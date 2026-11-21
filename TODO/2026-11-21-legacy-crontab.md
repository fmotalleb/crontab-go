# Legacy crontab configuration

- [x] Inspect the CLI, scheduler, and command runner.
- [x] Create the feature branch in an isolated worktree from origin/main.
- [x] Share the parser between direct config loading and YAML conversion.
- [x] Check randomized timing, calendar rules, environment scope, percent input, and direct/YAML execution parity.
- [x] Review the diff and run the build, vet, normal tests, race tests, and lint checks.

Use ordered NAME=value assignments to preserve environment names through Viper's YAML decoding. The feature targets Cronie file syntax with local execution and the application's existing output stream. The pinned Staticcheck analyzer panics in Go 1.27's internal/poll package on the base commit as well. The rest of the configured lint suite passes with Staticcheck disabled.
