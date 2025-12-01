# WASA Text – Chat Application (Course Project)

This repository contains the implementation of a full-stack chat system developed for the **Web and Software Architecture (WASA)** course  
taught by *Prof. Emanuele Panizzi* at *Sapienza University of Rome*.  

The project builds upon the official **"Fantastic Coffee (Decaffeinated)"** template provided in class and extends it into a working multi-user chat platform with RESTful APIs, SQLite storage, and an optional Vue-based WebUI.

---

## 1. Overview

The project’s objective is to design, implement, and deploy a web application following the principles of **modern software architecture**.  
The focus is on building a **RESTful API** backend, integrating it with a web frontend, and packaging the entire system into Docker containers.

**Key Learning Goals**
- Apply web architecture concepts (client–server model, REST APIs)
- Use standard web technologies (HTML, CSS, JS, Go, Vue.js)
- Implement data persistence using SQLite
- Manage software with Git and containerization tools (Docker)
- Deploy and test a complete web application

---

## 2. Technologies

| Component | Technology |
|------------|-------------|
| **Backend** | Go (Golang) |
| **Frontend** | Vue.js + Vite |
| **Database** | SQLite3 |
| **API Format** | REST / JSON |
| **Containerization** | Docker |
| **Documentation** | OpenAPI (YAML format) |

---

## 3. Project Structure

```

Wasa_proj/
├── cmd/webapi/             # Main backend executable
│   └── main.go             # Entry point for web server
├── service/api/            # REST API handlers (users, messages, groups)
├── service/database/       # SQLite database abstraction layer
├── service/models/         # Common data models
├── webui/                  # Optional Vue.js frontend
├── data/                   # SQLite database storage
├── uploads/                # Uploaded images and avatars
├── doc/                    # OpenAPI documentation and API specs
├── Dockerfile.backend      # Backend container build
├── Dockerfile.frontend     # Frontend container build
├── nginx.conf              # Proxy configuration (frontend → backend)
├── go.mod / go.sum         # Go module dependencies
└── README.md               # Project documentation

````

---

## 4. Setup and Execution

### Run Locally

```bash
go mod tidy
go run ./cmd/webapi --db-filename ./data/app.sqlite
````

Then visit: [http://localhost:3000](http://localhost:3000)

---

### API Test Example

```bash
# Check service health
curl http://localhost:3000/liveness

# Register or log in
curl -X POST http://localhost:3000/session \
  -H "Content-Type: application/json" \
  -d '{"name":"alice"}'
```

---

### Run with Docker

#### Backend

```bash
docker build -f Dockerfile.backend -t wasa-backend:latest .
docker run -d --name wasa-backend \
  -p 3000:3000 \
  -v $(pwd)/data:/data \
  -e WASA_DB_FILENAME=/data/app.sqlite \
  wasa-backend:latest
```

#### Frontend (Optional)

```bash
docker build -f Dockerfile.frontend -t wasa-frontend:latest .
docker run -d --name wasa-frontend -p 8080:80 wasa-frontend:latest
```

---

## 5. API Endpoints

| Endpoint                 | Method       | Description                    |
| ------------------------ | ------------ | ------------------------------ |
| `/liveness`              | GET          | Health check                   |
| `/session`               | POST         | Create or authenticate session |
| `/users/me`              | GET          | Retrieve current user profile  |
| `/conversations`         | GET / POST   | List or create conversations   |
| `/messages`              | GET / POST   | Retrieve or send messages      |
| `/messages/{id}`         | GET / DELETE | Get or delete specific message |
| `/messages/{id}/forward` | POST         | Forward a message              |
| `/messages/{id}/comment` | POST / GET   | Manage message comments        |

---

## 6. Development Notes

* The base architecture follows the **"Fantastic Coffee (Decaffeinated)"** template from the WASA course.
* The **API logic, database layer, and message-handling features** were implemented and extended independently.
* The frontend is built in **Vue.js (Vite)** and can be:

  * served standalone via Nginx, or
  * embedded in the backend binary (`-tags webui`).
* The project supports full containerization with Docker for both backend and frontend services.

---

## 7. Build and Deployment

### Development Mode

```bash
go run ./cmd/webapi
```

If using the web interface:

```bash
./open-node.sh
yarn run dev
```

### Production Mode

```bash
./open-node.sh
yarn run build-prod
yarn run preview
```

---

## 8. Known Limitations

* This system is designed **for educational purposes only**.
  It is not optimized for production use.
* Security features such as encryption and rate-limiting are minimal.
* Error handling and scalability are simplified to align with the course scope.

---

## 9. Credits

**Author:** Gioia Zheng
**Course:** Web and Software Architecture (WASA)
**Instructor:** Prof. Emanuele Panizzi
**University:** Sapienza University of Rome
**License:** MIT
**Based on:** *Fantastic Coffee (Decaffeinated)* project template
**Academic Year:** 2024–2025

---

```
