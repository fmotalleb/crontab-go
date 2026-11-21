# Preserve the Zap log level with OTLP

- [x] Trace logger initialization, the Zap tee, and the OTLP bridge.
- [x] Reproduce the level bypass with a regression test.
- [x] Make the OTLP core respect the configured Zap level.
- [x] Pass the YAML observability configuration first so the CLI activates the exporter.
- [x] Review, run the repository checks, and commit.

Zap tees enable a level if any core enables it. The OTLP bridge accepts levels below the local Zap threshold, so attaching it changes the logger's effective level and exports entries filtered out by the local sink. Use the existing core as the OTLP severity filter to preserve the configured level and keep the exporter's own filtering.

The CLI smoke test surfaced a startup issue in the pinned `gotel` integration. It reads configuration from the first option. Passing the version first discarded the YAML observability block and prevented the log bridge from being attached.

The regression reproduced the effective level changing from warn to debug before the fix. Build, vet, all tests, and the focused race test passed afterward. The CLI exported a warning through a loopback OTLP HTTP collector, suppressed its info log, and retained the configured file output, timestamp encoding, and caller setting.

Full lint hit a staticcheck crash in Go 1.27 standard-library analysis. The remaining configured linters reported zero issues.
