# WASA Text – Chat Application

This repository hosts the **WASA Text – Chat Application**, a WASA homework baseline built on top of the *Fantastic Coffee (decaffeinated)* starter shared in class. It is intended as a positive, educational full-stack chat project for practicing Web and Software Architecture concepts. The original fully fledged reference remains in the "Fantastic Coffee" repository.

## Overview
WASA Text is a client-server chat application with a Go backend and a Vue.js frontend. It exposes REST APIs for chat workflows and ships a Web UI for interactive usage, demos, and manual testing in a course setting.

## Tech stack & skills

**Backend**
- Go (modules + vendoring)
- HTTP API server
- Dependency injection / wiring in `cmd/`

**Frontend**
- Vue.js + Vite
- Bootstrap UI + Feather icons
- Yarn with offline mirror

**Skills emphasized**
- Web API design and documentation (OpenAPI)
- Full-stack integration (frontend consumes backend API)
- Build/release workflows with asset embedding

## Project structure
- `cmd/` — all executables. Go programs here should focus on executable concerns (CLI/env parsing, wiring, etc.).
  - `cmd/healthcheck/` — example daemon for checking the health of server processes; useful when the hypervisor lacks HTTP readiness/liveness probes (e.g., Docker engine).
  - `cmd/webapi/` — example web API server daemon.
- `demo/` — demo configuration file.
- `doc/` — documentation (for APIs, this is typically an OpenAPI file).
- `service/` — packages that implement project-specific functionality.
  - `service/api/` — example API server.
  - `service/globaltime/` — wrapper around `time.Time` (useful for unit testing).
- `vendor/` — managed by Go; contains vendored dependencies.
- `webui/` — Vue.js web frontend including:
  - Bootstrap UI framework
  - Customized "Bootstrap dashboard" template
  - Feather icons as SVG
  - Go code for release embedding

Other project files include:

- `open-node.sh` — starts a new (temporary) container using the `node:20` image for safe frontend development.

## What the Web UI is for
The Web UI (`webui/`) is a lightweight client used to:
- Log in/register users (if enabled in your API)
- Browse users and chat conversations
- Create and send messages
- Quickly verify API behavior without external tools

It is useful for demos, manual QA, and validating end-to-end API integration.

## Quickstart

```bash
git clone https://github.com/GioiaZheng/go-chat-system.git
cd go-chat-system

go mod download
go run ./cmd/webapi/
```

Expected output:

```text
API listening on 0.0.0.0:3000
```

For the Web UI:

```bash
cd webui
yarn install
yarn run dev
```

Expected output:

```text
VITE v...
Local: http://localhost:<port>/
```

Docker Compose:

```bash
docker compose up --build
```

Expected output:

```text
backend-1   | API listening on 0.0.0.0:3000
frontend-1  | /docker-entrypoint.sh: Configuration complete; ready for start up
```

Open the Web UI at `http://localhost:8080`; the API is exposed at `http://localhost:3000`.

## API tests

The API test suite covers authentication parsing, login behavior, user search, conversation membership checks, message access, and message mutation authorization.

```bash
go test ./service/api ./service/database
```

Expected output:

```text
ok  	github.com/GioiaZheng/Wasa_proj/service/api
ok  	github.com/GioiaZheng/Wasa_proj/service/database
```

## Security Notes

The current authentication model is documented in [docs/auth-threat-model.md](docs/auth-threat-model.md). It is intentionally lightweight for local development and coursework; production use should replace user-ID tokens with signed, expiring sessions.

## Screenshots

The documentation screenshots are stored with explicit filenames:

- [docs/screenshots/chat-login-screen.svg](docs/screenshots/chat-login-screen.svg)
- [docs/screenshots/chat-conversation-screen.svg](docs/screenshots/chat-conversation-screen.svg)

## Known limitations

This project is intentionally scoped as a course/full-stack homework baseline rather than a production-grade or research system. Current limitations include:

- **Simplified authentication** — local development uses a lightweight authentication model suitable for coursework, not a hardened session system.
- **Limited authorization hardening** — key chat flows include access checks, but the project has not been fully security-audited for production use.
- **SQLite/dev deployment** — the default database and deployment flow are optimized for local development and grading, not high-availability operation.
- **No formal research benchmark yet** — the project is useful for implementation practice and demos, but it does not include a validated research benchmark or comparative evaluation suite.

## Go vendoring

This project uses Go vendoring. After changing dependencies (`go get` or `go mod tidy`), run:

```bash
go mod vendor
```
Commit all files under the `vendor/` directory.

- More information: <https://go.dev/ref/mod#vendoring>
- Guidance: <https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html>

## Node/Yarn vendoring

The repository uses Yarn with an offline mirror to vendor dependencies. Commit the files inside the `.yarn` directory.

## How to customize for this chat project

1. Confirm the Go module path for your environment in `go.mod`, `go.sum`, and any `*.go` files.
2. Keep the API documentation current in `doc/api.yaml`, reflecting the chat endpoints and payloads.
3. If you do not need the WebUI for a given exercise, you can remove `webui/` and `cmd/webapi/register-webui.go`; otherwise keep them for the chat frontend.
4. Update the top-level package comment in `cmd/webapi/main.go` with the current project goal (WASA Text – Chat Application) and any environment notes.
5. Extend the `run()` function (`cmd/webapi/main.go`) to wire databases or external resources used by the chat service.
6. Implement chat-related API logic inside `service/api/` and add supporting packages under `service/` (or subdirectories).

## How to build

If you're **not** using the WebUI, or you don't want to embed the WebUI into the final executable:

```bash
go build ./cmd/webapi/
```

If you're using the WebUI **and** want to embed it into the final executable:

```bash
./open-node.sh
# inside the container
yarn run build-embed
exit
# outside the container
go build -tags webui ./cmd/webapi/
```

## How to run (development mode)

Launch the backend only:

```bash
go run ./cmd/webapi/
```
Launch the WebUI (in a new tab):

```bash
./open-node.sh
# inside the container
yarn run dev
```

## How to build for release / homework delivery

```bash
./open-node.sh
# inside the container
yarn run build-prod
```

### My build works with `yarn run dev`, but there is a JavaScript crash in production/grading

Some errors are not surfaced by Vite development mode. To preview the code that will be used in production/grading:

```bash
./open-node.sh
# inside the container
yarn run build-prod
yarn run preview
```

## Testing

The first `go test ./...` invocation compiles the vendored SQLite3 C driver. That
CGO build can take close to a minute and may look stalled even though it is
working. Once the driver object files are cached, subsequent `go test ./...`
runs finish in under a second.

Detailed testing notes, including Windows CGO troubleshooting and CI coverage,
are in [TESTING.md](TESTING.md).

## License

See [LICENSE](LICENSE).
