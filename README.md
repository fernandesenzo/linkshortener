# linkshortener

a small URL shortener API written in Go and using Redis.

it is intentionally simple: send an URL, get a short code, and use that code to redirect to the original URL.


## what it does

- creates short-lived links;
- redirects short codes to the original URLs;
- stores links in Redis;
- limits link creation per IP address;
- exposes a health endpoint for monitoring and deployments.

## project structure

```text
├── .github/workflows/   # CI pipeline configuration
├── cmd/api/             # application entrypoint
└── internal/
    ├── infra/           # redis client initialization
    ├── logger/           # logger setup using slog
    ├── middleware/       # recovery, request ID, logs, headers and body limits
    └── link/             # link-shortening domain logic
        ├── codegen/      # random code generator
        ├── handler/      # HTTP handlers
        ├── repository/   # redis data layer and rate limiting
        ├── service/      # application logic
        └── link.go       # domain configuration and core types
```

## running locally

copy the environment template:

```bash
cp .env.example .env
```

start the API and Redis with docker compose:

```bash
docker compose up --build
```

by default:

- the API listens on `SERVER_PORT` inside the container (`8080` in the example environment);
- docker exposes it on `localhost:HOST_PORT` (`8081` in the example environment);
- Redis is available to the API through `redis-local:6379`.

you can change the host facing API port using `HOST_PORT` in `.env`.

to run the API directly on your machine while keeping Redis in Docker:

```bash
docker compose up -d redis-local
REDIS_ADDR=localhost:6379 go run ./cmd/api
```

## API

### create a short link

```http
POST /links
Content-Type: application/json
```

Request:

```json
{
  "url": "https://example.com"
}
```

Successful response:

```http
201 Created
Content-Type: application/json
```

```json
{
  "code": "A8K2ZP"
}
```

the request must contain a JSON object with only the `url` field. URLs must use `http` or `https` and include a host.

possible errors:

| situation | status | response body |
| --- | ---: | --- |
| malformed JSON or unknown field | `400` | `invalid request` |
| invalid or empty URL | `400` | `invalid url` |
| unsupported content type | `415` | `unsupported content type` |
| URL exceeds the configured limit | `422` | `url too long` |
| IP reached its active-link limit | `422` | `ip already has the limit of links shortened, try again later.` |
| request body exceeds 4 KiB | `413` | `request body too large` |
| unexpected server or Redis error | `500` | `internal server error` |

error responses are currently plain text. successful link creation responses are JSON.

### redirect to the original URL

```http
GET /{code}
```

Example:

```bash
curl -i --max-redirs 0 http://localhost:8081/A8K2ZP
```

for a valid existing code, the API returns:

```http
307 Temporary Redirect
Location: https://example.com
```

the redirect is temporary (`307`), so clients preserve the original HTTP method when following it.

possible errors:

| situation | status | response body |
| --- | ---: | --- |
| code is not exactly six characters | `400` | `bad request` |
| code does not exist or has expired | `404` | `no link with this code` |
| unexpected server or Redis error | `500` | `internal server error` |

### health check

```http
GET /health
```

the endpoint performs a real Redis ping, so it verifies the dependency the API needs in order to work.

when the API and Redis are healthy:

```http
200 OK
Content-Type: application/json
```

```json
{
  "message": "ok"
}
```

when Redis is unavailable:

```http
503 Service Unavailable
Content-Type: text/plain; charset=utf-8
```

```text
error
```

## current limits and behavior

the current implementation intentionally has a small set of fixed rules:

- generated codes have six characters;
- codes use digits and uppercase letters (`0-9`, `A-Z`);
- URLs are valid only when they use `http` or `https` and contain a host;
- URLs are limited to 200 bytes;
- link-creation request bodies are limited to 4096 bytes;
- each IP can have up to 10 active links;
- links expire automatically after 24 hours;
- creating the same URL multiple times creates multiple different codes;
- collisions are retried a limited number of times;
- there is no authentication or ownership system;
- anyone with a valid code can use the redirect;
- there are no endpoints to list, update or delete links;
- the API does not check whether the destination URL is reachable.

the active-link limit depends on the client IP. when running behind a trusted proxy, the API reads `CF-Connecting-IP` or `X-Forwarded-For`. those headers should not be trusted when the API is publicly reachable without a trusted proxy in front of it.

## configuration note

some domain rules are currently represented by constants in `internal/link/link.go`:

- link expiration time;
- maximum active links per IP;
- maximum URL length;
- generated code length;
- maximum collision attempts.

changing those values currently requires changing the code and deploying a new version. moving some of these values to the env file application is one of the improvements planned for this project.


## tests and linter

run the test suite with the race detector:

```bash
go test -race -count=1 -shuffle=on ./...
```

run the linter:

```bash
golangci-lint run
```

build the application:

```bash
go build ./...
```

## possible next steps

some ideas that could make this a more complete service:

- configurable expiration time
- link deletion
- ownership and authentication
- click counters and basic analytics
- stronger abuse protection than IP-based limits
- an explicit choice between temporary and permanent redirects
