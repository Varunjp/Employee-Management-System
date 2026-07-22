# Employee Management System

A production-style Employee Management REST API built in Go, following
**Clean Architecture**. It supports full CRUD on employee records, backed
by PostgreSQL (raw `pgx` queries — no ORM), Redis caching for reads,
Swagger/OpenAPI documentation, and JWT-protected write endpoints. Fully
containerized with Docker Compose.

> See [ARCHITECTURE.md](./ARCHITECTURE.md) for a deep dive into the layering
> and design decisions.

## Tech Stack

| Concern            | Choice                                     |
|---------------------|---------------------------------------------|
| Language            | Go 1.25                                     |
| Web framework       | [Echo v4](https://echo.labstack.com/)       |
| Database            | PostgreSQL 16, via `jackc/pgx/v5` (raw SQL, no ORM) |
| Cache               | Redis 7, via `redis/go-redis/v9`            |
| API docs            | Swagger 2.0 via `swaggo/swag` + `swaggo/echo-swagger` |
| Auth                | JWT (`golang-jwt/jwt/v5`) on mutating endpoints |
| Tests               | Go standard `testing` + `testify`           |
| Containerization    | Docker multi-stage build + Docker Compose   |

## Project Structure

```
employee-management-system/
├── cmd/api/main.go                 # composition root / entrypoint
├── config/                         # env-based configuration loader
├── internal/
│   ├── domain/                     # entities, ports (interfaces), errors
│   ├── usecase/                    # business logic + unit tests
│   ├── repository/postgres/        # raw pgx SQL implementation
│   ├── cache/redis/                 # go-redis cache implementation
│   └── delivery/http/               # Echo routes, handlers, middleware
├── pkg/
│   ├── database/                   # postgres pool + redis client bootstrap
│   ├── auth/                       # JWT sign/verify helpers
│   └── logger/                     # minimal logging façade
├── docs/                           # Swagger spec (docs.go + swagger.json)
├── migrations/                     # versioned up/down SQL migrations
├── scripts/init-db.sql             # auto-run by the postgres container on first boot
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── .env.example
└── ARCHITECTURE.md
```

## Prerequisites

- **Docker & Docker Compose** (recommended path — no local Go/Postgres/Redis needed)
- *or*, for running natively: Go 1.22+, PostgreSQL 16+, Redis 7+

## Quick Start (Docker — recommended)

```bash
git clone <this-repo-url>
cd employee-management-system

cp .env.example .env
# Edit .env and set a real JWT_SECRET (e.g. `openssl rand -hex 32`)

make docker-up
# equivalent to: docker compose up --build -d
```

This starts three containers: `employee-api` (port `8080`), `employee-postgres`
(port `5432`), and `employee-redis` (port `6379`). The database schema is
created automatically on first boot via `scripts/init-db.sql`.

Check it's running:

```bash
curl http://localhost:8080/health
```

Open Swagger UI: **http://localhost:8080/swagger/index.html**

Stop the stack:

```bash
make docker-down          # keeps the postgres volume (data persists)
make docker-clean         # also wipes the postgres volume
```

## Running Natively (without Docker)

1. Start your own PostgreSQL and Redis instances.
2. Apply the schema:
   ```bash
   psql "$DATABASE_URL" -f migrations/000001_create_employees_table.up.sql
   ```
3. Copy and edit the env file:
   ```bash
   cp .env.example .env
   ```
4. Fetch dependencies and run:
   ```bash
   go mod tidy
   make run          # or: go run ./cmd/api
   ```

## Environment Variables

All variables are documented in [`.env.example`](./.env.example).

| Variable                  | Default            | Description                                   |
|----------------------------|---------------------|------------------------------------------------|
| `APP_PORT`                 | `8080`               | HTTP port the API listens on                    |
| `DB_HOST` / `DB_PORT`       | `localhost` / `5432` | PostgreSQL connection host/port                |
| `DB_USER` / `DB_PASSWORD`   | `postgres` / `postgres` | PostgreSQL credentials                    |
| `DB_NAME`                   | `employee_db`        | Database name                                  |
| `DB_SSLMODE`                | `disable`            | pgx sslmode                                    |
| `REDIS_ADDR`                | `localhost:6379`     | Redis host:port                                |
| `REDIS_PASSWORD`            | *(empty)*            | Redis password, if any                         |
| `REDIS_DB`                  | `0`                  | Redis logical DB index                         |
| `JWT_SECRET`                | *(required)*         | HMAC secret used to sign JWTs                  |
| `JWT_EXPIRATION_MINUTES`    | `60`                 | Access token lifetime                          |
| `ADMIN_USERNAME`/`ADMIN_PASSWORD` | `admin`/`admin123` | Demo credentials for `/auth/login`       |

## API Overview

Base path: `/api/v1`. Full interactive documentation is served at
`/swagger/index.html` once the app is running.

| Method | Path                     | Auth required | Description                     |
|--------|--------------------------|---------------|----------------------------------|
| POST   | `/auth/login`            | No            | Exchange credentials for a JWT   |
| POST   | `/employees`             | **Yes**       | Create an employee                |
| GET    | `/employees`             | No            | List all employees (cached)      |
| GET    | `/employees/{id}`        | No            | Get one employee (cached)        |
| PUT    | `/employees/{id}`        | **Yes**       | Update an employee                |
| DELETE | `/employees/{id}`        | **Yes**       | Delete an employee                |
| GET    | `/health` (root path)    | No            | Liveness check                    |

### 1. Authenticate

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Response:

```json
{
  "access_token": "eyJhbGciOi...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

Use the token in the `Authorization: Bearer <token>` header for
create/update/delete requests below.

### 2. Create an employee

```bash
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
        "name": "John Doe",
        "position": "Software Engineer",
        "salary": 60000,
        "hired_date": "2024-06-01"
      }'
```

```json
{
  "id": 1,
  "name": "John Doe",
  "position": "Software Engineer",
  "salary": 60000,
  "hired_date": "2024-06-01",
  "created_at": "2024-06-10T12:00:00Z"
}
```

### 3. Get an employee

```bash
curl http://localhost:8080/api/v1/employees/1
```

### 4. Update an employee

```bash
curl -X PUT http://localhost:8080/api/v1/employees/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
        "name": "John Doe",
        "position": "Senior Software Engineer",
        "salary": 80000,
        "hired_date": "2024-06-01"
      }'
```

### 5. Delete an employee

```bash
curl -i -X DELETE http://localhost:8080/api/v1/employees/1 \
  -H "Authorization: Bearer $TOKEN"
# -> 204 No Content
```

### 6. List employees

```bash
curl http://localhost:8080/api/v1/employees
```

## Caching Strategy

`GET /employees` and `GET /employees/{id}` follow a **cache-aside**
pattern:

1. On read, check Redis first (`employee:{id}` or `employees:all`).
2. On a cache miss, fetch from PostgreSQL and populate Redis with a 5-minute TTL.
3. On create/update/delete, the affected keys are proactively invalidated
   so subsequent reads never serve stale data.

Cache read/write failures are logged but never fail the request — Redis
is a performance optimization, not a source of truth.

## Error Handling

Every non-2xx response returns a consistent JSON envelope:

```json
{ "message": "employee not found" }
```

| Situation                          | HTTP status |
|-------------------------------------|-------------|
| Validation failure (bad input)      | 400         |
| Missing/invalid/expired JWT         | 401         |
| Employee ID not found               | 404         |
| Unexpected/infrastructure error     | 500         |

## Testing

```bash
make test          # go test ./... -race -count=1
make test-cover    # generates coverage.html
```

Unit tests for business logic (`internal/usecase`) use hand-rolled
in-memory fakes for the repository and cache ports — no database or
network is required to run them.

## Continuous Integration

`.github/workflows/ci.yml` runs on every push/PR to `main`: it spins up
Postgres and Redis service containers, applies the migration, runs
`go vet`, a `gofmt` check, the full test suite with the race detector,
builds the binary, and finally builds the Docker image.

## Regenerating Swagger Docs

The `docs/` package is checked in and works out of the box. If you
change any `@swagger`-annotated handler or the `router.go` general API
comment, regenerate it with:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swag
```

## License

MIT
