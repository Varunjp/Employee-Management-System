# Architecture

This service follows **Clean Architecture**. Dependencies only ever point
inward, toward the domain. Outer layers know about inner layers; inner
layers know nothing about outer layers.

```
                       ┌─────────────────────────────┐
                       │   cmd/api/main.go            │   composition root
                       │   (wires everything together)│
                       └───────────────┬──────────────┘
                                        │ constructs & injects
                 ┌──────────────────────┼──────────────────────┐
                 ▼                                              ▼
┌────────────────────────────────┐                 ┌─────────────────────────────┐
│ internal/delivery/http          │                 │ internal/repository/postgres │
│  - router.go (Echo routes)      │                 │  - raw pgx SQL queries        │
│  - handler/ (HTTP <-> DTOs)     │                 │                              │
│  - middleware/ (JWT, errors)    │                 │ internal/cache/redis          │
└───────────────┬─────────────────┘                 │  - go-redis client            │
                 │ implements                        └───────────────┬──────────────┘
                 ▼ domain.EmployeeUsecase                             │ implements
┌─────────────────────────────────┐                                  ▼ domain.EmployeeRepository /
│ internal/usecase                │◄─────────────depends on───────── domain.Cache
│  - business rules                │
│  - cache-aside orchestration     │
└───────────────┬──────────────────┘
                 │ depends only on
                 ▼
┌─────────────────────────────────┐
│ internal/domain                  │   <-- the center. No framework,
│  - Employee entity                │       driver, or transport import.
│  - EmployeeRepository (port)      │
│  - Cache (port)                   │
│  - EmployeeUsecase (port)         │
│  - sentinel errors                │
└───────────────────────────────────┘
```

## Layers

### `internal/domain` — Entities & Ports
The innermost layer. Contains the `Employee` entity, DTOs used across
layer boundaries, sentinel errors (`ErrEmployeeNotFound`, etc.), and the
**interfaces** (ports) that outer layers must implement:

- `EmployeeRepository` — persistence port
- `Cache` — caching port
- `EmployeeUsecase` — the port the delivery layer calls into

This package imports nothing from the rest of the module and nothing
framework-specific. It has zero knowledge of HTTP, Postgres, or Redis.

### `internal/usecase` — Business Logic
Implements `domain.EmployeeUsecase`. Contains all business rules:
input validation, cache-aside read/write/invalidate orchestration, and
error propagation. It depends only on the `domain.EmployeeRepository` and
`domain.Cache` **interfaces** — never on `pgx` or `go-redis` directly. This
is what makes `internal/usecase/employee_usecase_test.go` able to test
business logic with simple in-memory fakes and zero network calls.

### `internal/repository/postgres` & `internal/cache/redis` — Infrastructure Adapters
Concrete implementations of the domain ports.
- `postgres.employeeRepository` talks to PostgreSQL using **raw SQL via
  `pgx`** (no ORM, per project requirements).
- `redis.cache` talks to Redis via `go-redis`.

Either could be swapped (e.g. Postgres → MySQL, Redis → in-memory LRU)
by writing a new adapter that implements the same domain interface,
without touching `internal/usecase` or `internal/domain`.

### `internal/delivery/http` — Transport / Presentation
Echo HTTP handlers, routes, and middleware. Handlers translate JSON
requests into domain input types, call the usecase, and translate domain
results/errors back into JSON responses and HTTP status codes. This is
also where JWT authentication and Swagger documentation are wired in.

### `config` & `pkg`
- `config` loads and validates environment variables into a typed struct.
- `pkg/database` bootstraps the Postgres pool and Redis client (with
  bounded startup retries so container start order doesn't matter).
- `pkg/auth` contains JWT sign/parse helpers shared by the login handler
  and the auth middleware.
- `pkg/logger` is a tiny logging façade so no other package imports the
  standard `log` package directly.

### `cmd/api/main.go` — Composition Root
The **only** file that imports every concrete implementation. It builds
the dependency graph (repository → cache → usecase → handlers → router)
and starts the HTTP server with graceful shutdown on `SIGINT`/`SIGTERM`.

## Why this shape?

- **Testability** — business rules in `internal/usecase` are tested with
  plain Go structs, no database, no HTTP server, no mocking framework.
- **Replaceability** — Postgres, Redis, and Echo are all detail choices
  that live behind interfaces; none of them leak into `domain` or `usecase`.
- **Single Responsibility per package** — a bug in caching logic, SQL,
  or HTTP routing is isolated to one small package.

## Request lifecycle example: `GET /api/v1/employees/{id}`

1. `router.go` routes the request to `EmployeeHandler.GetByID`.
2. The handler parses and validates the `id` path parameter.
3. The handler calls `EmployeeUsecase.GetByID(ctx, id)`.
4. The usecase checks `Cache.Get("employee:{id}")`.
   - **Cache hit:** unmarshal and return immediately.
   - **Cache miss:** call `EmployeeRepository.GetByID`, populate the
     cache with a TTL, and return the result.
5. The handler converts the `domain.Employee` to `domain.EmployeeResponse`
   and writes JSON.
6. If any step returns a domain sentinel error (e.g.
   `ErrEmployeeNotFound`), it propagates up to the centralized
   `middleware.NewHTTPErrorHandler`, which maps it to the correct HTTP
   status code and a consistent `{"message": "..."}` JSON body.
