# Short URL

A URL shortener backend written in Go. It exposes a small REST API to register users, log in with JWT, and turn long URLs into 7-character short codes. Guests can create short codes too; logged-in users are identified through an optional auth middleware.

## Features

- Short code generation (7 chars, base62, cryptographically secure via `crypto/rand`)
- User registration and login with JWT (HS256) bearer tokens
- Optional authentication on short code creation (guests allowed, bad tokens rejected)
- Protected `/user/me` endpoint
- Request ID middleware (`X-Request-ID` is reused if sent, generated otherwise)
- Structured JSON logging with `log/slog`
- PostgreSQL with a `pgx` connection pool
- Database migrations with `golang-migrate`
- Layered architecture: handler → service → repository

## Tech Stack

| Area       | Library                                  |
| ---------- | ---------------------------------------- |
| HTTP       | `net/http` (Go 1.22+ `ServeMux` routing) |
| Database   | PostgreSQL, `jackc/pgx/v5` (`pgxpool`)   |
| Auth       | `golang-jwt/jwt/v5`                      |
| Migrations | `golang-migrate/migrate/v4`              |
| Config     | `joho/godotenv`                          |
| IDs        | `google/uuid`                            |

## Project Structure

```
├── bin
│   └── main
├── cmd
│   ├── main
│   │   └── main.go
│   └── migration
│       └── migration.go
├── go.mod
├── go.sum
├── internal
│   ├── auth                                      # creating access token and vcalidate it
│   │   └── token.go
│   ├── config                                    # checks all config is present or not
│   │   └── config.go
│   ├── database                                  # postgresql connection
│   │   └── database.go
│   ├── handler                                   # api gateway
│   │   ├── health.go
│   │   ├── url_handler.go
│   │   └── user_handler.go
│   ├── httpx                                     # custom error handler
│   │   └── error.go
│   ├── middleware                                # to add requestId and check is user authenticated or not
│   │   ├── auth.go
│   │   └── requestID.go
│   ├── model                                     # user and url tables
│   │   ├── url_model.go
│   │   └── user_model.go
│   ├── repo                                       # all daatabase query
│   │   ├── url_repository.go
│   │   └── user_repository.go
│   ├── security                                    # encode and decode the password
│   │   └── password.go
│   └── service # all business logic
│       ├── short_code.go
│       ├── url_service.go
│       └── user_service.go
├── Makefile
└── migrations                                       # all database operations
    ├── 0001_create_table_urls.down.sql
    ├── 0001_create_table_urls.up.sql
    ├── 0002_create_table_user.down.sql
    ├── 0002_create_table_user.up.sql
    ├── 0003_add_foreign_key.down.sql
    └── 0003_add_foreign_key.up.sql
```

## Getting Started

### Prerequisites

- Go 1.22 or newer
- PostgreSQL

### 1. Clone

```bash
git clone https://github.com/Sahilkumar121/short_url.git
cd short_url
```

### 2. Configure environment

Create a `.env` file in the project root. All three values are required; the app exits on startup if any is missing.

```env
PORT=8080
DATABASE=your-database-url
SECRETKEY=replace-with-a-long-random-secret
```

| Variable    | Description                         |
| ----------- | ----------------------------------- |
| `PORT`      | Port the server listens on          |
| `DATABASE`  | PostgreSQL connection string        |
| `SECRETKEY` | Secret used to sign and verify JWTs |

### 3. Run migrations

The migration runner reads files from the `migrations/` directory.

```bash
make migrate-up      # apply all migrations
make migrate-down    # roll back one step
```

### 4. Start the server

```bash
make run
```

The API is then available at `http://localhost:<PORT>/api/v1`.

## API Reference

Base path: `/api/v1`

| Method | Endpoint                    | Auth     | Description                        |
| ------ | --------------------------- | -------- | ---------------------------------- |
| GET    | `/healthz`                  | None     | Health check                       |
| POST   | `/user/register`            | None     | Create an account                  |
| POST   | `/user/login`               | None     | Log in and receive a JWT           |
| GET    | `/user/me`                  | Required | Get the current user               |
| POST   | `/url/shortcode`            | Optional | Create a short code for a long URL |
| GET    | `/url/longurl/{short_code}` | None     | Resolve a short code to its URL    |

### Register

`POST /api/v1/user/register`

```json
{
  "name": "Sahil",
  "email": "sahil@example.com",
  "password": "your-password"
}
```

Responses: `201 Created`, `400 Bad Request`, `409 Conflict` (user already exists)

```json
{ "message": "register successful" }
```

### Login

`POST /api/v1/user/login`

```json
{
  "email": "sahil@example.com",
  "password": "your-password"
}
```

Responses: `200 OK`, `400 Bad Request`, `403 Forbidden` (invalid credentials), `404 Not Found`

```json
{
  "message": "loging successful",
  "token": "<jwt>",
  "token-type": "Bearer"
}
```

Tokens expire after **15 minutes**.

### Get current user

`GET /api/v1/user/me`

```
Authorization: Bearer <jwt>
```

```json
{
  "id": "b1c2d3e4-...",
  "name": "Sahil",
  "email": "sahil@example.com"
}
```

### Create a short code

`POST /api/v1/url/shortcode`

The `Authorization` header is optional. With no header the request is treated as a guest; with an invalid or expired token it is rejected with `401`.

```json
{ "long_url": "https://example.com/some/very/long/path" }
```

```json
{ "short_code": "aB3xK9Z" }
```

### Resolve a short code

`GET /api/v1/url/longurl/{short_code}`

```json
{ "long_url": "https://example.com/some/very/long/path" }
```

Returns `404 Not Found` if the code does not exist.

## Example Usage

```bash
# register
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Sahil","email":"sahil@example.com","password":"secret123"}'

# login
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"email":"sahil@example.com","password":"secret123"}'

# create a short code (add -H "Authorization: Bearer <token>" to create it as a user)
curl -X POST http://localhost:8080/api/v1/url/shortcode \
  -H "Content-Type: application/json" \
  -d '{"long_url":"https://example.com/some/very/long/path"}'

# resolve it
curl http://localhost:8080/api/v1/url/longurl/aB3xK9Z
```

## Authentication Details

- Algorithm: HS256, with the signing method enforced during validation
- Issuer: `my-app` (checked on every request)
- Expiry: 15 minutes, and a token without an expiry is rejected
- Subject (`sub`): the user's UUID, placed into the request context by the middleware

## Server Configuration

| Setting       | Value      |
| ------------- | ---------- |
| Read timeout  | 1 minute   |
| Write timeout | 1 minute   |
| Idle timeout  | 30 seconds |
| DB max conns  | 25         |
| DB conn idle  | 30 minutes |
| DB conn life  | 1 hour     |

## Roadmap

- [ ] Redirect endpoint (`GET /{short_code}` → `302`)
- [ ] Click analytics
- [ ] Short code expiry and custom aliases
- [ ] Rate limiting
- [ ] Refresh tokens
- [ ] Dockerfile and docker-compose
- [ ] Tests and CI

## License

Add a license of your choice (for example MIT).
