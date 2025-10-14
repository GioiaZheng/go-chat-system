---

# WASA Text — Chat API & WebUI

A lightweight chat platform built with **Go**, **SQLite**, and an optional **embedded WebUI**.
It supports users, private and group conversations, messages with pagination, forwarding, comments, and more — all via clean RESTful APIs.

---

## Features

✅ RESTful backend written in Go
✅ SQLite-based persistent storage
✅ JWT-style lightweight session auth
✅ Supports private & group conversations
✅ Message features: pagination, forward, comment, delete
✅ WebUI embeddable in backend (with `-tags webui`)
✅ Fully containerized — ready for Docker deployment

---

## Project Structure

```
Wasa_proj/
├── cmd/webapi/             # Main backend entrypoint
├── service/api/            # HTTP handlers (users, messages, groups, etc.)
├── service/database/       # SQLite database layer
├── service/models/         # Shared data models
├── webui/                  # Optional frontend (Vue/Vite)
├── data/                   # SQLite storage (mounted in Docker)
├── uploads/                # Uploaded files (avatars, images)
├── Dockerfile.backend      # Backend image (multi-stage build)
├── Dockerfile.frontend     # Frontend (optional, Nginx-based)
├── nginx.conf              # Frontend proxy config
├── go.mod / go.sum         # Dependencies
└── README.md               # You're here
```

---

## Quick Start (Local Development)

### Install dependencies

```bash
go mod tidy
```

### Run backend

```bash
go run ./cmd/webapi --db-filename ./data/app.sqlite
```

### Test API

```bash
# Check health
curl -s http://localhost:3000/liveness | jq

# Register/login (creates token)
TOKEN=$(curl -s -X POST http://localhost:3000/session \
  -H 'Content-Type: application/json' \
  -d '{"name":"alice"}' | jq -r '.data.token')

# Create a conversation
CID=$(curl -s -X POST http://localhost:3000/conversations \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"demo","member_ids":[]}' | jq -r '.data.conversation.id')

# Send a message
curl -s -X POST http://localhost:3000/messages \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"conversationId\":\"$CID\",\"content\":\"hello from API\"}" | jq

# Retrieve messages
curl -s "http://localhost:3000/messages?conversationId=$CID&limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## Run with Docker (Backend only)

### Build image

```bash
docker build -f Dockerfile.backend -t wasa-backend:latest .
```

### Run container

```bash
mkdir -p ./data

docker run -d --name wasa-backend \
  -p 3000:3000 \
  -v $(pwd)/data:/data \
  -e WASA_DB_FILENAME=/data/app.sqlite \
  wasa-backend:latest
```

### Verify

```bash
curl -s http://localhost:3000/liveness | jq
```

---

## Build Backend + Embedded WebUI

If you want your backend binary to also serve the WebUI (no need for Nginx):

```bash
docker build -f Dockerfile.backend -t wasa-backend:webui \
  --build-arg BUILD_TAGS=webui .
docker run -d --name wasa-backend-webui \
  -p 3000:3000 \
  -v $(pwd)/data:/data \
  wasa-backend:webui
```

Then open:
👉 [http://localhost:3000](http://localhost:3000)

---

## Frontend (Optional, standalone)

You can also deploy the `webui` separately using Nginx:

```bash
docker build -f Dockerfile.frontend -t wasa-frontend:latest .
docker run -d --name wasa-frontend -p 8080:80 wasa-frontend:latest
```

Then visit [http://localhost:8080](http://localhost:8080)
(Default Nginx proxy config will forward `/api/` to backend.)

---

## API Overview

| Endpoint                   | Method     | Description                      |
| -------------------------- | ---------- | -------------------------------- |
| `/liveness`                | GET        | Health check                     |
| `/session`                 | POST       | Login or register user           |
| `/users/me`                | GET        | Get own profile                  |
| `/conversations`           | GET/POST   | List or create conversations     |
| `/messages`                | GET/POST   | Retrieve or send messages        |
| `/messages/{id}`           | GET/DELETE | Get or delete a specific message |
| `/messages/{id}/forward`   | POST       | Forward a message                |
| `/messages/{id}/comment`   | POST       | Add a comment                    |
| `/messages/{id}/comment`   | GET        | Get comments                     |
| `/messages/{id}/uncomment` | POST       | Remove all comments              |

---

## Environment Variables

| Variable           | Default            | Description                   |
| ------------------ | ------------------ | ----------------------------- |
| `WASA_DB_FILENAME` | `/data/app.sqlite` | SQLite database path          |
| `GIN_MODE`         | `release`          | Run mode                      |
| `PORT`             | `3000`             | Server port                   |
| `WITH_WEBUI`       | optional           | Build tag for embedded web UI |

---

## Housekeeping

To clean local environment:

```bash
docker stop wasa-backend && docker rm wasa-backend
docker image prune -f
docker builder prune -f
```

---

## Author & Credits

**Project Maintainer:** Gioia Zheng
**Backend Language:** Go 1.22
**Database:** SQLite3
**Frontend:** Vue + Vite
**License:** MIT

---

## Example Output

```bash
$ curl -s http://localhost:3000/liveness | jq
{
  "code": 200,
  "message": "Service is alive"
}

$ curl -s http://localhost:3000/messages?conversationId=$CID | jq
{
  "code": 200,
  "message": "Messages retrieved successfully",
  "data": {
    "messages": [
      {
        "id": "d7b...",
        "content": "hello from API",
        "senderId": "u_1760...",
        "conversationId": "a4d5...",
        "createdAt": "2025-10-14T14:10:33Z"
      }
    ]
  }
}
```

---

## Final Notes

* You can run the backend **alone** or **embed** the frontend using `-tags webui`.
* SQLite keeps all data in `/data/app.sqlite`. Mount it in Docker to persist.
* Code structure follows clean layering:
  `api → service/database → models`
* For debugging, use:

  ```bash
  go run ./cmd/webapi --db-filename ./data/app.sqlite --verbose
  ```