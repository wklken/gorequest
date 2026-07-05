# AGENTS.md

## Project Overview

GoRequest is a Go HTTP client library inspired by Node.js SuperAgent. The module path is `github.com/wklken/gorequest`.

This is a single-package repository. Production code and tests live at the repository root:

- Core request API: `gorequest.go`
- Constants and helpers: `constant.go`, `util.go`, `logger.go`, `stats.go`
- Tests: `gorequest_test.go`, `util_test.go`

## Setup Commands

- Download dependencies: `go mod download`
- Refresh module metadata when dependencies change: `go mod tidy`
- Sync a vendor tree only when intentionally working with vendored dependencies: `make vendor`

The module declares `go 1.21` in `go.mod`. GitHub Actions tests Go 1.21.x through 1.26.x.

## Development Workflow

- There is no development server; this is a library package.
- Build/check compilation: `go test ./...`
- Run the CI-equivalent test command: `go test -v ./...`
- Run the Makefile test target: `make test`
- Format changed Go files with `gofmt -s -w <files>`.

Keep changes small and local. Public API behavior is concentrated in `SuperAgent` methods in `gorequest.go`; read nearby tests before changing request, header, body, cookie, proxy, retry, clone, or mock behavior.

## Testing Instructions

- Run all tests as CI does: `go test -v ./...`
- Run all tests without verbose logs: `go test ./...`
- Run one test: `go test -run TestName ./...`
- Run package coverage: `go test ./... -covermode=count -coverprofile .coverage.cov`

The Makefile target `make test` runs:

```sh
go test ./... -covermode=count -coverprofile .coverage.cov
```

Use `make vendor` only when a task explicitly needs a vendor tree.

Tests use Go's standard `testing` package, `httptest`, `gock`, and `goproxy`. Add or update tests in the existing root-level `_test.go` files for behavior changes.

## Code Style

- Keep the package name `gorequest`.
- Match existing Go style and naming, including current exported API names.
- Use `gofmt -s`; do not hand-format imports or whitespace.
- Add GoDoc comments for new exported functions, types, variables, or constants.
- Avoid introducing new dependencies when the standard library or existing dependencies are enough.
- Do not reformat unrelated code or modernize deprecated standard-library usage unless the task asks for it.

The configured linter is `golangci-lint` via `.golangci.yaml`. If linting is required, run `make lint`; install the pinned tool with `make init`.

## Build and CI

GitHub Actions runs on pushes and pull requests to `master` and `develop` with a Go version matrix from 1.21.x through 1.26.x:

```sh
go test -v ./...
```

There is no separate build artifact or deployment process for this library.

## Pull Request Guidelines

- Include tests for new or changed behavior.
- Run `gofmt` on changed Go files.
- Run `go test -v ./...` before submitting code changes.
- Update `README.md` when adding or changing user-facing functionality.
- Commit subjects should be imperative, capitalized, and have no trailing period, following `CONTRIBUTING.md`.

## Gotchas

- `README.md` and `CONTRIBUTING.md` include older upstream references, but this checkout's module path is `github.com/wklken/gorequest`.
- Some tests emit large verbose logs. Use non-verbose `go test ./...` when you only need pass/fail feedback.
- `make init` installs the pinned `golangci-lint` version from the Makefile.
