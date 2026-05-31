# Testing

This project uses Go tests for the backend and Yarn-based checks for the Web
UI and OpenAPI specification.

## Recommended Environment

The most reliable test environment is Linux, WSL, or the GitHub Actions Ubuntu
runner. The backend depends on `github.com/mattn/go-sqlite3`, which compiles a
CGO SQLite driver during test and build steps.

Required tools:

- Go 1.22 or newer
- CGO-enabled C compiler
- Node.js 20
- Corepack / Yarn

## Backend Tests

Run the focused API and database tests:

```bash
go test ./service/api ./service/database
```

Run the full backend test suite:

```bash
go test ./...
```

The first run can be slow because Go compiles the SQLite C driver. Later runs
reuse the build cache.

## Windows Notes

On Windows, `go-sqlite3` requires a compatible MinGW-w64 toolchain. Linker
errors mentioning libraries such as `mingw32`, `mingwex`, `pthread`, or
`kernel32` usually point to a local CGO toolchain mismatch rather than an
application test failure.

If this happens, use one of these paths:

```bash
# WSL or Linux
go test ./...
```

or rely on GitHub Actions, which runs the suite on Ubuntu.

## Frontend and API Checks

Install frontend dependencies:

```bash
corepack enable
yarn install --cwd webui --immutable
```

Build the frontend:

```bash
yarn --cwd webui run build-prod
```

Lint the OpenAPI specification:

```bash
yarn install --immutable
yarn exec spectral lint doc/api.yaml --ruleset doc/spectral.yaml
```

## CI Coverage

The GitHub Actions workflow runs:

- Go formatting check,
- `go test ./...`,
- `go vet ./...`,
- Web UI dependency install,
- Web UI production build,
- OpenAPI linting.
