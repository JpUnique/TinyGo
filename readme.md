# TinyGo — URL Shortener Backend (Go)

TinyGo is a **production-ready URL shortener backend** written in **Go**, designed as a system‑design–driven project. It demonstrates clean architecture, scalability patterns, Dockerized infrastructure, database migrations, caching, and background workers.

This project is ideal for:

* Backend engineering interviews
* DevOps / Cloud practice
* System design learning
* Portfolio demonstration

---

## ✨ Features

* 🔗 Create short URLs (custom or auto‑generated)
* 🚀 Fast redirects with Redis caching
* 📊 Click counting (Redis + Postgres)
* 🧱 Clean layered architecture
* 🐳 Docker & docker‑compose ready
* 🗄️ PostgreSQL with versioned migrations
* ⚙️ Background worker for analytics flush
* 🔐 Production‑safe timeouts & graceful shutdown

---

## 🏗️ Architecture Overview

```
Client
  ↓
HTTP API (Chi Router)
  ↓
Service Layer (business logic)
  ↓
Store Layer
  ├── PostgreSQL (persistent data)
  └── Redis (cache + counters)
```

### Key Design Choices

* **Postgres** = source of truth
* **Redis** = cache + click aggregation
* **pgxpool** = high‑performance DB access
* **golang-migrate** = schema versioning
* **Distroless runtime image** = security

---

## 📂 Project Structure

```
tinygo/
├── cmd/
│   └── server.go              # App bootstrap & HTTP server
├── pkg/
│   ├── api/                   # HTTP handlers
│   ├── service/               # Business logic
│   ├── store/                 # Postgres & Redis access
│   ├── worker/                # Background jobs
│   ├── middleware/            # Logging, recovery
│   └── utils/                 # Helpers (random ID)
├── migrations/                # SQL migrations
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

---

## ⚙️ Environment Variables

| Variable       | Description     | Example                                                             |
| -------------- | --------------- | ------------------------------------------------------------------- |
| `PORT`         | API port        | `8080`                                                              |
| `BASE_URL`     | Public base URL | `http://localhost:8080`                                             |
| `POSTGRES_URL` | Postgres DSN    | `postgres://postgres:postgres@postgres:5432/tinygo?sslmode=disable` |
| `REDIS_ADDR`   | Redis address   | `redis:6379`                                                        |

---

## 🐳 Running with Docker (Recommended)

### Start everything (DB, Redis, migrations, API)

```bash
docker-compose up --build
```

Services started:

* PostgreSQL
* Redis
* Migration runner
* TinyGo API

API will be available at:

```
http://localhost:8080
```

---

## 🔁 Database Migrations

Migrations live in the `migrations/` folder.

Example:

```
0001_init.up.sql
0001_init.down.sql
```

### Apply migrations

Automatically run by docker-compose.

### Run manually

```bash
docker-compose run migrate up
docker-compose run migrate down 1
```

---

## 📡 API Endpoints

### ➕ Create short URL

`POST /shorten`

```json
{
  "long_url": "https://example.com",
  "custom": "myalias"
}
```

Response:

```json
{
  "short_url": "http://localhost:8080/myalias"
}
```

---

### 🔀 Redirect

`GET /{code}`

Example:

```
GET /abc123
```

Redirects to original URL and increments click count.

---

## 🧠 Background Worker

A background worker periodically:

* Scans Redis click counters
* Flushes aggregated counts into PostgreSQL

This avoids heavy DB writes on every request.

---

## 🧪 Testing (Planned)

* Integration tests with Docker
* Redis & Postgres test containers
* Load testing scripts

---

## 🔐 Production Considerations

* Redis used as best‑effort cache
* DB always source of truth
* Graceful shutdown enabled
* Distroless runtime image
* Context timeouts everywhere

---

## 🚀 Future Improvements

* URL expiration
* User accounts & auth
* Rate limiting
* Analytics dashboard
* Kubernetes deployment
* Terraform infrastructure

---

## 👨‍💻 Author

Built by **JpUnique** as a marathon backend & DevOps learning project.

---

## 📜 License

MIT License
