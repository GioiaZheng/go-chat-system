# WASA Text – Chat Application

This repository hosts the **WASA Text – Chat Application**, built on top of the *Fantastic Coffee (decaffeinated)* starter shared in class. The goal is to provide a clean baseline for the Web and Software Architecture (WASA) homework while keeping production concerns out of scope. The original fully fledged reference remains in the "Fantastic Coffee" repository.

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
- `webui/` — example Vue.js web frontend including:
  - Bootstrap JavaScript framework
  - Customized "Bootstrap dashboard" template
  - Feather icons as SVG
  - Go code for release embedding

Other project files include:

- `open-node.sh` — starts a new (temporary) container using the `node:20` image for safe frontend development.

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

## How to build for production / homework delivery

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

## License

See [LICENSE](LICENSE).
