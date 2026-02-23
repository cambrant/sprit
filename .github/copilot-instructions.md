# CLAUDE.md

## Build & Run

Always use the Makefile commands below to build, run, and test the game. They set necessary environment variables and ensure the correct working directory. Don't run `go build` or `go test` directly, unless when working on a specific package in `internal/` — even then, prefer the Makefile for consistency.

```
make          # build the game binary
make test     # run all tests
make run      # build and run
make race     # run tests with race detector
make coverage # generate HTML coverage report
```

WASM: `make build-wasm` then `make serve-wasm` (localhost:8080).

## Codebase Guide

**Read `docs/ARCHITECTURE.md` first.** It maps every file to its responsibility, documents game modes, data flow, async systems, and the full directory layout. Consult it instead of running find/grep or reading all files — it is the definitive guide to the codebase.

When making changes, keep `docs/ARCHITECTURE.md` up to date so it stays accurate. If you add, remove, rename, or change the responsibility of any file or package, update the architecture doc to match.

## Development Roadmap & Bugs

The development roadmap lives in `docs/ROADMAP.md`. It contains version history, planned features, and an **Active bug list** at the bottom. Keep it current:

- When completing a roadmap item, check its box.
- When adding new features or fixing bugs, add or update the relevant entries.
- When fixing a bug from the active bug list, check its box.

## Testing

Run tests using `make test`. Tests are located in `_test.go` files throughout the codebase. They use Go's standard `testing` package. For more complex tests, see `internal/testutils/` for helper functions and test data.

When adding new features, also add tests to cover them. When fixing bugs, add tests that reproduce the bug before fixing it, then verify the tests pass after the fix.