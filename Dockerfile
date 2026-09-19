FROM golang:1.26.8-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

FROM alpine:3.24

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder --chown=root:appgroup /app/main .

RUN chmod 750 /app/main

USER appuser

EXPOSE 8080

CMD ["./main"]