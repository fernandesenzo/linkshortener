a simple URL shortener API written in Go using redis.

## project structure

```
├── .github/workflows/   # CI pipeline configuration
├── cmd/api/             # app entrypoint (main.go)
└── internal/            
    ├── infra/           # redis client initialization
    ├── logger/          # logger setup (slog)
    ├── middleware/      # global HTTP middlewares (recovery, request ID, logs)
    └── link/            # link-shortening domain logic
        ├── codegen/     # short code generator (base62)
        ├── handler/     # HTTP handlers
        ├── repository/  # redis data layer (CRUD + rate-limits)
        ├── service/     # orchestrator / service layer
        └── link.go      # Domain configurations and core structs
```

## running
1. copy the environment template:
   ```bash
   cp .env.example .env
   ```

2. start the services using docker dompose:
   ```bash
   docker compose up --build
   ```
   the API will run on port `8081`, but you can change it to your favorite port on docker-compose.yml.

you can also run the api locally:
```bash
docker compose up -d redis-local
go run ./cmd/api
```

### link creation
`POST /api/links`
```json
{
  "url": "https://example.com"
}
```

response:
```json
{
  "code": "e9a71f"
}
```

### redirect
`GET /{code}`

redirects (`307`) to the original URL.

## business logic
by default, codes have length 6, last for 24 hours and each IP address can have 10 links created simoultaneously, you can change these settings and some others on internal/link/link.go

## tests and linter

run tests:
```bash
go test ./...
```

run linter:
```bash
golangci-lint run
```

there is also a github actions pipeline to run tests and linter on every push to main

