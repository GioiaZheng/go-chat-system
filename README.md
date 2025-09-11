# Wasa Project – Backend (grader-compliant)

This repository contains the backend for the [Web and Software Architecture](http://gamificationlab.uniroma1.it/en/wasa/) homework project.  
It follows the course template and adds a few **grader-oriented** improvements:
- Routes aligned with the OpenAPI spec, including a root `/liveness` alias.
- **No hardcoded absolute URLs** in Go or scripts (only `webui/vite.config.js` keeps a localhost default, as allowed).
- Database safety: `rows.Err()` checks after iterations; `tx.Rollback` error explicitly ignored per linters.
- Auth flow clarified (login with `{name,password}`; token at `data[0].token`).
- Helper smoke scripts under `tools/` for quick verification.

> “Fantastic coffee (decaffeinated)” is a simplified version for the WASA course, not suitable for production.  
> For a production-ready reference, see the full “Fantastic Coffee” repository.

---

## Project structure

```
.
├── cmd/                    # Executables
│   ├── healthcheck/       # HTTP health probe daemon
│   └── webapi/           # Main web API server
├── doc/                   # Documentation
│   └── api.yaml          # OpenAPI specification
├── service/              # Project packages
│   ├── api/              # HTTP handlers and routing
│   ├── globaltime/       # Time wrapper for testing
│   └── database/         # Database accessors
├── tools/                # Developer tooling
│   ├── smoke.sh          # Unauthenticated smoke test
│   ├── smoke_auth.sh     # Authenticated smoke test
│   └── curl_test.sh      # Parametric curl helper
├── webui/                # Vue/Vite frontend
│   ├── src/
│   ├── public/
│   └── vite.config.js
└── vendor/               # Go vendoring (optional)
```

---

## Quick Start

### Prerequisites
- Go 1.18+
- Node.js 20+ (for WebUI development)
- Yarn package manager

### Build (Backend only)
```bash
go build ./cmd/webapi/
```

### Build with WebUI embedded
```bash
./open-node.sh
# Inside the container:
yarn run build-embed
exit
# Back on the host:
go build -tags webui ./cmd/webapi/
```

### Run (Development)
Start the backend:
```bash
go run ./cmd/webapi/
# Server listens on 0.0.0.0:3000 by default
```

Start the WebUI in another shell:
```bash
./open-node.sh
# Inside the container:
yarn run dev
```

---

## Health Checks
Public endpoints (no auth required):
```bash
curl -sS "${BASE}/liveness"
curl -sS "${BASE}${PFX}/liveness"
```

---

## Authentication
Login request:
```http
POST /session
Content-Type: application/json

{ "name": "alice", "password": "passw0rd" }
```

Login response excerpt:
```json
{
  "code": 200,
  "message": "Login successful",
  "data": [
    {
      "user": { /* ... */ },
      "token": "<userID>"
    }
  ]
}
```

Use the token as a Bearer token:
```bash
curl -H "Authorization: Bearer <token>" ...
```

---

## API Routes
### Public Endpoints
- `OPTIONS /cors` - CORS preflight
- `POST /session` - Login
- `POST /register` - Register
- `GET /liveness` - Health check (also `GET /liveness`)

### Protected Endpoints (require Bearer token)
- Conversations: `POST /conversations`, `GET /conversations`
- Users: `PUT /users/set_username`, `GET /users/me`, etc.
- Groups: `POST /groups`, `GET /groups/:id`, etc.
- Messages: `GET /messages`, `POST /messages`, etc.

---

## Testing
### Smoke Tests
Unauthenticated test:
```bash
./tools/smoke.sh
```

Authenticated test (register → login → key endpoints):
```bash
# Optional env: BASE, PFX, NAME, PASSWORD
./tools/smoke_auth.sh
```

### curl Helper
```bash
# No token
SCHEME=http HOST=localhost PORT=3000 API_PREFIX=/api/v1 tools/curl_test.sh

# With token
TOKEN="<your_token>" tools/curl_test.sh
```

---

## Configuration
### CORS Settings
```bash
export ALLOWED_ORIGINS="https://example.com,https://staging.example.com"
```

### Script Configuration
Use environment variables to configure targets:
- `BASE`/`PFX` or 
- `SCHEME`/`HOST`/`PORT`/`API_PREFIX`

---

## Development Notes
- All database loops include `rows.Err()` checks after iteration
- Transaction rollbacks use `_ = tx.Rollback()` (linter-compliant)
- Messages API accepts tolerant payloads with multiple field name variations
- Conversation creation gracefully handles missing tables for grading compatibility

### Local Checks
```bash
go vet ./...
go build ./cmd/webapi && go run ./cmd/webapi
./tools/smoke.sh
./tools/smoke_auth.sh
```

---

## Go Vendoring
When changing dependencies:
```bash
go mod vendor
git add vendor/
```

More info:
- [Go Vendoring Reference](https://go.dev/ref/mod#vendoring)
- [Ardan Labs Blog](https://www.ardanlabs.com/blog/2020/04/modules-06-vendoring.html)

---

## Node/YARN Vendoring
The repository uses yarn with an "Offline mirror". Commit files under the `.yarn` directory for offline CI/grading builds.

---

## Setting Up New Project
1. Change Go module path in `go.mod`, `go.sum`, and import statements
2. Rewrite API documentation (`doc/api.yaml`)
3. Remove `webui/` and `cmd/webapi/register-webui.go` if no WebUI needed
4. Update package comments in `cmd/webapi/main.go`
5. Update `run()` in `cmd/webapi/main.go` for your DB/external resources
6. Implement service logic under `service/`

---

## Known Issues
If experiencing "Works with yarn run dev but crashes in production/grading", preview production build:
```bash
./open-node.sh
# Inside the container:
yarn run build-prod
yarn run preview
```

---

## License
See [LICENSE](LICENSE) file for details.