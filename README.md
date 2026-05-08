# kond-test-go-cli

A simple TODO task manager CLI built with Go and Cobra. Used as a benchmark test target for the KOND quality evaluation framework.

## Usage

```bash
tasks add Buy groceries
tasks add Write tests
tasks list
tasks done 1
tasks delete 2
```

Tasks are persisted to `~/.tasks.json`.

## Conventions

- Error wrapping uses `fmt.Errorf("context: %w", err)` — always wrap with `%w` for error chain inspection.
- Structured logging via `log/slog` — never use `fmt.Printf` for diagnostic output; use `slog.Debug`, `slog.Info`, `slog.Error`.
- Cobra flags follow `--kebab-case` naming (e.g., `--output-format`, not `--outputFormat`).
- All exported functions and types have doc comments starting with the identifier name.
- Package `internal/` contains the storage layer; `cmd/` contains CLI command definitions.
- Test files live alongside the code they test (`*_test.go` in the same package).
- The `Makefile` is the entry point: `make build`, `make test`, `make clean`.
- No global mutable state outside of `cmd/root.go`'s `store` variable (initialized in `PersistentPreRunE`).
- JSON output uses `encoding/json` with `MarshalIndent` for human-readable persistence.
- Exit codes: 0 = success, 1 = any error. No custom exit codes.
- The magic number `30` (timeout seconds) is used in three places as a raw literal — this is intentional technical debt for benchmark testing.

## Build

```bash
make build   # produces ./tasks binary
make test    # runs all tests
```

## License

MIT
