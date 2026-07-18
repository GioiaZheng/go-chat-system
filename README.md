# WASA Text

[![CI](https://github.com/GioiaZheng/go-chat-system/actions/workflows/ci.yml/badge.svg)](https://github.com/GioiaZheng/go-chat-system/actions/workflows/ci.yml)
[![Secret scan](https://github.com/GioiaZheng/go-chat-system/actions/workflows/secret-scan.yml/badge.svg)](https://github.com/GioiaZheng/go-chat-system/actions/workflows/secret-scan.yml)

A full-stack chat application with a Go REST API, Vue frontend, SQLite
persistence, OpenAPI documentation, authorization checks, and reproducible local
deployment.

This is a Web and Software Architecture course project. It demonstrates
end-to-end system integration and security-aware API design; it is not presented
as a production messaging service.

## What it implements

- user registration, login, and user search;
- one-to-one and group conversation workflows;
- message creation, access, and mutation rules;
- conversation membership and authorization checks;
- a Vue + Vite interface for interactive use and manual QA;
- an OpenAPI contract under [`doc/api.yaml`](doc/api.yaml);
- SQLite-backed local persistence;
- Docker Compose for a two-service local deployment;
- automated API, vulnerability, and secret-scanning checks.

## Architecture

```text
Vue + Vite web client
        │
        │ HTTP / JSON
        ▼
Go REST API
        │
        ▼
SQLite persistence
```

The backend keeps executable wiring under `cmd/` and project logic under
`service/`. The frontend consumes the documented API and can run separately
during development or be embedded into the Go binary for release builds.

## Quickstart

### Backend

Requires Go and a working CGO toolchain for the SQLite driver.

```bash
git clone https://github.com/GioiaZheng/go-chat-system.git
cd go-chat-system

go mod download
go run ./cmd/webapi/
```

Expected startup message:

```text
API listening on 0.0.0.0:3000
```

### Frontend

```bash
cd webui
yarn install
yarn run dev
```

Vite prints the local development URL after startup.

### Docker Compose

```bash
docker compose up --build
```

Open the Web UI at `http://localhost:8080`; the API is exposed at
`http://localhost:3000`.

## Testing

The API suite covers authentication parsing, login behavior, user search,
conversation membership, message access, and message mutation authorization.

```bash
go test ./service/api ./service/database
```

For the full repository suite:

```bash
go test ./...
```

The first full run compiles the vendored SQLite3 C driver and can take close to
a minute. Subsequent runs use the build cache. See [TESTING.md](TESTING.md) for
Windows CGO troubleshooting and CI details.

## API and security model

The API contract is maintained in [`doc/api.yaml`](doc/api.yaml).

The authentication model is intentionally lightweight and suitable for local
development and coursework. It uses user-ID tokens rather than signed,
expiring sessions. The assumptions, trust boundaries, and production
hardening requirements are documented in
[docs/auth-threat-model.md](docs/auth-threat-model.md).

Automated repository checks include:

- Go tests and build validation;
- `govulncheck` for known Go dependency vulnerabilities;
- secret scanning;
- frontend build checks in CI.

These checks improve the development baseline but do not constitute a
production security audit.

## SQLite connection policy

The API opens SQLite once at startup and shares a bounded `database/sql` pool
across requests. The pool is limited to eight open and four idle connections;
foreign-key enforcement, a five-second busy timeout, and WAL journaling are
applied through driver options to every physical connection. Query result sets
are closed before dependent follow-up queries, avoiding pool self-deadlocks.

This is a correctness-oriented local deployment policy, not a published
throughput claim. The repository does not currently report QPS or P99 latency
without a reproducible benchmark and workload definition.

## Screenshots

- [Login screen](docs/screenshots/chat-login-screen.svg)
- [Conversation screen](docs/screenshots/chat-conversation-screen.svg)

## Repository layout

```text
cmd/
  webapi/          API server and dependency wiring
  healthcheck/     process health-check utility
service/
  api/             HTTP handlers and chat workflows
  database/        persistence layer
  globaltime/      testable time abstraction
webui/             Vue + Vite frontend
doc/api.yaml       OpenAPI contract
docs/              security, testing, and screenshots
demo/              local demo configuration
vendor/            vendored Go dependencies
```

## Build modes

Backend without an embedded frontend:

```bash
go build ./cmd/webapi/
```

Frontend development through the provided Node container:

```bash
./open-node.sh
# inside the container
yarn run dev
```

Release build with embedded frontend assets:

```bash
./open-node.sh
# inside the container
yarn run build-embed
exit
go build -tags webui ./cmd/webapi/
```

To preview the production frontend bundle before delivery:

```bash
./open-node.sh
# inside the container
yarn run build-prod
yarn run preview
```

## Known limitations

- Authentication is designed for coursework, not hardened production use.
- Authorization tests cover key chat flows, but the system has not undergone a
  full security audit.
- SQLite and the default deployment target local development rather than
  high-availability operation.
- The project does not include a formal performance or research benchmark.
- The dependency workflow uses Go vendoring and a Yarn offline mirror, which
  increases repository size but supports reproducible course builds.

## Course context and attribution

WASA Text was developed for the Web and Software Architecture course using the
*Fantastic Coffee (decaffeinated)* starter shared in class.

The public history begins with
[`bd2d40e`](https://github.com/GioiaZheng/go-chat-system/commit/bd2d40ebd1cc1e8670b8b0607e00575b2bf35705),
a full project upload rather than an untouched copy of the starter. It therefore
does not support file-by-file authorship claims. Attribution is deliberately
stated at the feature and development-phase level:

| Phase | Scope |
|---|---|
| Course starter | Initial full-stack project scaffold and WASA build conventions. The original starter snapshot is not present in this repository, so those foundations are not claimed as original work. |
| Course application | Chat-specific user, one-to-one and group conversation, message, reply, forward, and deletion workflows; SQLite-backed application state; the OpenAPI contract; Vue integration; and local Docker packaging. |
| Later repository maintenance | API and authorization tests, CI, secret and vulnerability scanning, security documentation, dependency maintenance, and SQLite connection and query-lifecycle hardening. |

The later-maintenance boundary is visible in the history: CI
([`c3afd7f`](https://github.com/GioiaZheng/go-chat-system/commit/c3afd7f)), API
tests ([`9b70ea4`](https://github.com/GioiaZheng/go-chat-system/commit/9b70ea4)),
secret and vulnerability scans
([`2376b4f`](https://github.com/GioiaZheng/go-chat-system/commit/2376b4f),
[`75ecfcf`](https://github.com/GioiaZheng/go-chat-system/commit/75ecfcf)), and
SQLite hardening
([`e61c6a6`](https://github.com/GioiaZheng/go-chat-system/commit/e61c6a6)).

## Dependency maintenance

After changing Go dependencies:

```bash
go mod tidy
go mod vendor
```

Commit the updated files under `vendor/`. The frontend similarly keeps a Yarn
offline mirror under `.yarn/`.

## License

See [LICENSE](LICENSE).
