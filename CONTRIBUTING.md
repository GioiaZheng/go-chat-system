# Contributing

Thanks for your interest in this project. This guide covers local setup for
the backend and Web UI, the checks that run in CI, and the conventions for
branches, commits, and pull requests.

## Setup

The project has a Go backend (`cmd/webapi/`) and a Vue Web UI (`webui/`).
Dependencies are vendored: Go modules under `vendor/`, and Yarn with an offline
mirror under `.yarn/`.

Backend:

```bash
go run ./cmd/webapi/
```

Web UI (separate terminal):

```bash
cd webui
yarn install --immutable
yarn run dev
```

The Web UI is served at `http://localhost:8080` and the API at
`http://localhost:3000`. The full stack also runs under Docker:

```bash
docker compose up --build
```

See [TESTING.md](TESTING.md) for the recommended environment (the SQLite driver
compiles with CGO, so a C compiler is required) and the exact tool versions.

## Checks

These run in CI and should pass locally before opening a pull request:

```bash
gofmt -l .            # must report no files
go test ./...
go vet ./...
cd webui && yarn run build-prod
yarn exec spectral lint doc/api.yaml --ruleset doc/spectral.yaml
yarn check:api-routes  # API routes match the OpenAPI spec
```

## Branches

Create a short-lived branch off `main` for each change and open a pull request
back into `main`. Use a short, descriptive branch name (for example
`fix/login-redirect` or `docs/security-policy`).

## Commits

- One logical change per commit.
- Subject line in the imperative mood and capitalized, e.g.
  `Add presence indicator to conversation list`.
- Keep the backend, Web UI, and OpenAPI spec consistent in the same change —
  if a route changes, update `doc/api.yaml` so `check:api-routes` stays green.

## Pull requests

- Open the pull request against `main`.
- CI (Go format/test/vet, frontend build, OpenAPI lint, route coverage) must be
  green before merge.
- Reference the issue the change closes in the description.

## Security

Do not report vulnerabilities through public issues or pull requests; see
[SECURITY.md](SECURITY.md).
