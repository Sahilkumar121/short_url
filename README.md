# Short URL

A URL shortener backend written in Go. It exposes a REST API to register users, authenticate with JWT, and convert long URLs into 7-character short codes.

Guests can create short URLs without authentication. Logged-in users are identified through optional authentication middleware, allowing their URLs to be associated with their accounts.

The application uses PostgreSQL for persistent storage and Redis for caching URL redirection data. It also implements asynchronous email notifications using SMTP and Go channels.

## Features

- **Short code generation:** 7-character Base62 codes generated securely using `crypto/rand`.
- **User registration and login:** Password hashing and JWT authentication using HS256.
- **Optional authentication:** Guests can create short URLs, while invalid or expired tokens are rejected.
- **Protected user endpoint:** Retrieve authenticated user details through `/user/me`.
- **Request ID middleware:** Reuses an incoming `X-Request-ID` or generates a new one.
- **Structured logging:** JSON logging using Go's `log/slog`.
- **PostgreSQL:** Persistent storage using `pgx` and `pgxpool`.
- **Database migrations:** Schema management using `golang-migrate`.
- **Asynchronous email notifications:** Sends login notification emails through an SMTP worker implemented using Go channels.
- **Redis caching:** Caches resolved URLs to reduce repeated PostgreSQL queries.
- **Layered architecture:** Separates HTTP handlers, business logic, and database operations.

## Tech Stack

| Area | Library / Technology |
|---|---|
| HTTP | `net/http` (Go 1.22+ ServeMux routing) |
| Database | PostgreSQL, `jackc/pgx/v5` (`pgxpool`) |
| Authentication | `golang-jwt/jwt/v5` |
| Password hashing | Argon2id |
| Migrations | `golang-migrate/migrate/v4` |
| Configuration | `joho/godotenv` |
| IDs | `google/uuid` |
| Caching | Redis, `redis/go-redis/v9` |
| Email | SMTP (Mailtrap for development) |
| Background processing | Go channels and goroutines |
| Logging | `log/slog` |

## Project Structure

```text
├── bin
│   └── main
├── cmd
│   ├── main
│   │   └── main.go
│   └── migration
│       └── migration.go
├── go.mod
├── go.sum
├── internal
│   ├── auth
│   │   └── token.go
│   ├── cache
│   │   └── redis.go
│   ├── config
│   │   └── config.go
│   ├── database
│   │   └── database.go
│   ├── email
│   │   └── SMTP email implementation
│   ├── handler
│   │   ├── health.go
│   │   ├── url_handler.go
│   │   └── user_handler.go
│   ├── httpx
│   │   └── error.go
│   ├── middleware
│   │   ├── auth.go
│   │   └── requestID.go
│   ├── model
│   │   ├── url_model.go
│   │   └── user_model.go
│   ├── repo
│   │   ├── url_repository.go
│   │   └── user_repository.go
│   ├── security
│   │   └── password.go
│   ├── service
│   │   ├── short_code.go
│   │   ├── url_service.go
│   │   └── user_service.go
│   └── worker
│       └── Email worker implementation
├── Makefile
└── migrations
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
- Redis
- SMTP credentials (Mailtrap or another SMTP provider)
- Make

### 1. Clone the repository

```bash
git clone https://github.com/Sahilkumar121/short_url.git
cd short_url
```

### 2. Configure environment variables

Create a `.env` file in the project root.

```env
PORT=8080

# PostgreSQL
DATABASE=your-postgresql-connection-string

# JWT
SECRETKEY=replace-with-a-long-random-secret

# SMTP
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USERNAME=your-smtp-username
SMTP_PASSWORD=your-smtp-password

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASS=
REDIS_DB=0
```

| Variable | Description |
|---|---|
| `PORT` | HTTP server port |
| `DATABASE` | PostgreSQL connection string |
| `SECRETKEY` | Secret used to sign and validate JWTs |
| `SMTP_HOST` | SMTP server hostname |
| `SMTP_PORT` | SMTP server port |
| `SMTP_USERNAME` | SMTP authentication username |
| `SMTP_PASSWORD` | SMTP authentication password |
| `REDIS_ADDR` | Redis server address |
| `REDIS_PASS` | Redis password (optional for local Redis) |
| `REDIS_DB` | Redis database index |

For Neon PostgreSQL, use the connection string provided by Neon, including its required SSL parameters.

For local Redis, the default address is `localhost:6379`. When running the application inside Docker, use the Redis service hostname instead.

**Important:** Add `.env` to `.gitignore`. Never commit database passwords, SMTP credentials, or JWT secrets.

### 3. Start Redis

You can run Redis using Docker:

```bash
docker run --name short-url-redis \
  -p 6379:6379 \
  -d redis:7-alpine
```

Check whether Redis is responding:

```bash
docker exec -it short-url-redis redis-cli ping
```

Expected output:

```text
PONG
```

### 4. Run database migrations

The migration runner reads SQL files from the `migrations/` directory.

```bash
make migrate-up
```

To roll back the most recent migration:

```bash
make migrate-down
```

### 5. Start the server

```bash
make run
```

The API is available at:

```text
http://localhost:8080/api/v1
```

## Email Notification Implementation

The application implements asynchronous email notifications using SMTP, Go channels, and goroutines.

Instead of sending an email directly inside the login HTTP request, the application uses a background email worker.

### How it works

1. A user submits their login credentials.
2. The service validates the user's email and password.
3. When authentication succeeds, the service submits a login notification to the email worker.
4. The worker processes the notification through a Go channel.
5. The SMTP handler sends the email to the user.
6. Email errors are logged without changing the successful login result.

### Architecture

```text
Login Request
     |
     v
User Handler
     |
     v
User Service
     |
     +---- Validate Credentials
     |
     v
Login Notification
     |
     v
Email Worker
     |
     v
Go Channel
     |
     v
SMTP Handler
     |
     v
User Email
```

The email worker is started during application startup and closed when the application shuts down.

This approach keeps SMTP operations out of the main login request flow and separates notification processing from authentication business logic.

## Redis Caching Implementation

Redis is integrated as an in-memory cache for URL resolution.

PostgreSQL remains the primary storage system, while Redis stores frequently accessed short-code mappings to reduce database queries.

### How URL caching works

When a client requests a long URL using a short code:

1. The application checks Redis for the requested short code.
2. If the data exists in Redis, the cached long URL is returned.
3. If the data is not found in Redis, the application queries PostgreSQL.
4. Once PostgreSQL returns the URL, the application stores the mapping in Redis.
5. The cached entry expires automatically after 30 minutes.

### Cache flow

```text
Client Request
      |
      v
Get Long URL
      |
      v
Check Redis
      |
      +------ Cache Hit ------> Return Long URL
      |
      |
      +------ Cache Miss
                    |
                    v
              PostgreSQL
                    |
                    v
               Get Long URL
                    |
                    v
               Store in Redis
                    |
                    v
              Return Long URL
```

### Redis configuration

The application uses:

```go
redisClient := redis.NewClient(&redis.Options{
    Addr:     cfg.Redis.Addr,
    Password: cfg.Redis.Password,
    DB:       cfg.Redis.RedisDB,
})
```

The cache TTL is configured as:

```go
redisHandler := cache.NewRedisCache(
    redisClient,
    30*time.Minute,
)
```

The cache uses a key prefix to identify URL redirection entries:

```text
shortCode:redirect:{short_code}
```

### Benefits of Redis

- Reduces repeated PostgreSQL queries for frequently accessed URLs.
- Provides fast in-memory lookups.
- Automatically expires cached entries through TTL.
- Keeps persistent URL data in PostgreSQL.

A cache miss falls back to PostgreSQL. Redis connection and command failures are currently surfaced by the URL service, so transparent fallback during a Redis outage is a future improvement.

## API Reference

Base path:

```text
/api/v1
```

| Method | Endpoint | Authentication | Description |
|---|---|---|---|
| GET | `/healthz` | None | Health check |
| POST | `/user/register` | None | Register a new user |
| POST | `/user/login` | None | Login and receive JWT |
| GET | `/user/me` | Required | Retrieve current user |
| POST | `/url/shortcode` | Optional | Create a short code |
| GET | `/url/longurl/{short_code}` | None | Resolve short code to long URL |

### Register

`POST /api/v1/user/register`

Request:

```json
{
  "name": "Sahil",
  "email": "sahil@example.com",
  "password": "your-password"
}
```

Responses: `201 Created`, `400 Bad Request`, `409 Conflict`.

Successful response:

```json
{
  "message": "register successful"
}
```

### Login

`POST /api/v1/user/login`

Request:

```json
{
  "email": "sahil@example.com",
  "password": "your-password"
}
```

Responses: `200 OK`, `400 Bad Request`, `403 Forbidden`, `404 Not Found`.

Successful response:

```json
{
  "message": "login successful",
  "token": "<jwt>",
  "token-type": "Bearer"
}
```

Tokens expire after 15 minutes.

A login notification is submitted to the asynchronous email worker after successful authentication.

### Get current user

`GET /api/v1/user/me`

Authorization header:

```text
Authorization: Bearer <jwt>
```

Response:

```json
{
  "id": "b1c2d3e4-...",
  "name": "Sahil",
  "email": "sahil@example.com"
}
```

### Create a short code

`POST /api/v1/url/shortcode`

The Authorization header is optional.

Without a token, the request is treated as a guest request. With a valid token, the URL can be associated with the authenticated user. Invalid or expired tokens are rejected with `401 Unauthorized`.

Request:

```json
{
  "long_url": "https://example.com/some/very/long/path"
}
```

Response:

```json
{
  "short_code": "aB3xK9Z"
}
```

### Resolve a short code

`GET /api/v1/url/longurl/{short_code}`

Example:

```text
GET /api/v1/url/longurl/aB3xK9Z
```

Response:

```json
{
  "long_url": "https://example.com/some/very/long/path"
}
```

Returns `404 Not Found` if the short code does not exist.

Redis is checked before PostgreSQL to avoid unnecessary database queries.

## Example Usage

### Register

```bash
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Sahil","email":"sahil@example.com","password":"secret123"}'
```

### Login

```bash
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"email":"sahil@example.com","password":"secret123"}'
```

### Create a short code

```bash
curl -X POST http://localhost:8080/api/v1/url/shortcode \
  -H "Content-Type: application/json" \
  -d '{"long_url":"https://example.com/some/very/long/path"}'
```

To create the URL as an authenticated user, add:

```bash
-H "Authorization: Bearer <token>"
```

### Resolve a short code

```bash
curl http://localhost:8080/api/v1/url/longurl/aB3xK9Z
```

## Authentication Details

- Algorithm: HS256.
- Signing method is enforced during token validation.
- Issuer: `my-app`.
- Expiry: 15 minutes.
- Tokens without an expiration claim are rejected.
- Subject (`sub`): User UUID.
- The middleware stores the authenticated user's ID in the request context.

## Server Configuration

| Setting | Value |
|---|---|
| Read timeout | 1 minute |
| Write timeout | 1 minute |
| Idle timeout | 30 seconds |
| Shutdown timeout | 10 seconds |
| DB max connections | 25 |
| DB connection idle time | 30 minutes |
| DB connection lifetime | 1 hour |
| Redis cache TTL | 30 minutes |

## Roadmap

- [ ] Redirect endpoint (`GET /{short_code}` → `302`)
- [ ] Click analytics
- [ ] Short code expiry and custom aliases
- [ ] Rate limiting
- [ ] Refresh tokens
- [ ] More resilient Redis outage fallback
- [ ] Dockerfile and docker-compose
- [ ] Unit and integration tests
- [ ] CI/CD pipeline

## License

Add a license of your choice (for example, MIT).
