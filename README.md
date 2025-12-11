# WASA Text – Chat Application

A full-stack chat system built for the *Web and Software Architecture (WASA)* course at Sapienza University of Rome. The project exposes a REST API written in Go, persists data with SQLite, and ships with an optional Vue 3 web interface that can be served separately or embedded in the backend binary.

## Features
- RESTful API for sessions, users, conversations, and messages.
- SQLite-backed persistence with automatic schema migrations on startup.
- Optional Vue 3 + Vite web interface (can be embedded with Go build tags).
- Dockerfiles for backend and frontend images, plus Nginx configuration for proxying.

## Requirements
- Go **1.17+**
- Node.js **18+** and Yarn **3.x** (for the web UI)
- SQLite3 (shared library used by the Go `sqlite3` driver)
- Docker (optional, for container builds)

## Repository layout
```
Wasa_proj/
├── cmd/webapi/        # Backend entrypoint and configuration
├── service/           # API handlers, database layer, and models
├── webui/             # Vue 3 single-page app (Vite)
├── doc/               # OpenAPI document (api.yaml) and Spectral config
├── Dockerfile.backend # Backend image definition
├── Dockerfile.frontend# Frontend image definition
├── nginx.conf         # Example reverse-proxy for the frontend
└── README.md
```

## Backend quick start
1. Fetch dependencies:
   ```bash
   go mod tidy
   ```
2. Start the API server (listens on `0.0.0.0:3000` by default):
   ```bash
   CFG_DB_FILENAME=./data/wasatext.db \
   go run ./cmd/webapi
   ```
3. Verify health:
   ```bash
   curl http://localhost:3000/liveness
   ```

### Configuration
Configuration values can come from environment variables (prefix `CFG_`), CLI flags, or a YAML file (`/conf/config.yml` by default). Key options:

| Name | Default | Notes |
| ---- | ------- | ----- |
| `CFG_WEB_APIHOST` | `0.0.0.0:3000` | API listen address |
| `CFG_WEB_DEBUGHOST` | `0.0.0.0:4000` | Debug/pprof listener |
| `CFG_DB_FILENAME` | `/tmp/decaf.db` | SQLite database file |
| `CFG_CONFIG_PATH` | `/conf/config.yml` | Optional config file |

Example YAML (`config.yml`):
```yaml
config:
  path: ./config.yml
web:
  apihost: 0.0.0.0:3000
  readtimeout: 5s
  writetimeout: 5s
db:
  filename: ./data/wasatext.db
```

## Web UI
The Vue 3 frontend lives in `webui/`.

- Install dependencies (Yarn 3 is already vendored via `yarn.lock`):
  ```bash
  cd webui
  yarn install
  ```
- Run in development mode (defaults to Vite dev server on port 5173):
  ```bash
  yarn dev
  ```
- Build a static bundle (served by Nginx or any static host):
  ```bash
  yarn build-prod
  ```

### Embedding the web UI into the Go binary
1. Build the embedded assets (outputs to `webui/dist`):
   ```bash
   cd webui
   yarn build-embed
   ```
2. Build or run the backend with the `webui` tag to serve the SPA from `/` while keeping API routes intact:
   ```bash
   go run -tags webui ./cmd/webapi
   # or
   go build -tags webui -o webapi ./cmd/webapi
   ./webapi
   ```

## Docker
### Backend image
```bash
docker build -f Dockerfile.backend -t wasa-backend:latest .
docker run -d --name wasa-backend \
  -p 3000:3000 \
  -v $(pwd)/data:/data \
  -e CFG_DB_FILENAME=/data/app.sqlite \
  wasa-backend:latest
```

### Frontend image
```bash
docker build -f Dockerfile.frontend -t wasa-frontend:latest .
docker run -d --name wasa-frontend -p 8080:80 wasa-frontend:latest
```

## API reference
The OpenAPI specification lives at `doc/api.yaml`. Some notable routes:

- `GET /liveness` – service health check
- `POST /session` – create or authenticate a session
- `GET /users/me` – current user profile
- `GET/POST /conversations` – list or create conversations
- `GET/POST /messages` – fetch or send messages

## Development notes
- Use `go test ./...` for backend unit tests. If you enable the embedded UI, ensure `webui/dist` exists before running tests/builds.
- The backend logs to stdout and responds to `SIGTERM`/`SIGINT` with graceful shutdown.
- Default ports: API `3000`, debug/pprof `4000`, Vite dev server `5173`, example Nginx frontend `8080`.

## Credits
- Author: Gioia Zheng
- Course: Web and Software Architecture (WASA) — Prof. Emanuele Panizzi
- University: Sapienza University of Rome
- License: MIT
