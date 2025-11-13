# Chirpy

A Twitter-like REST API clone built with Go and PostgreSQL for posting short messages with authentication, refresh tokens, and premium user features. This is an example project for learning and reference.

## Features

- JWT-based authentication with refresh token rotation
- Create, read, and delete chirps (140 character limit)
- Built-in profanity filter
- Premium user upgrades (Chirpy Red) via webhooks
- User account management

## Installation and Setup

### Prerequisites

- Go 1.25+
- PostgreSQL (or Docker)
- [goose](https://github.com/pressly/goose) - database migrations
- [sqlc](https://sqlc.dev/) - type-safe SQL code generation

Install tools:

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Setup

1. **Start PostgreSQL**

```sh
docker compose up -d
```

2. **Configure environment** - Create `.env` file:

```
DB_URL="postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"
PLATFORM="dev"
JWT_SECRET="your-secret-key"
POLKA_KEY="your-polka-key"
```

3. **Run migrations**

```sh
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" -dir sql/schema/ up
```

4. **Build and run**

```sh
go build -o out && ./out
```

Server runs on `http://localhost:8080`
