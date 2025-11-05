# Chirpy

## Build and run

```sh
go build -o out && ./out
```

## Run postgres docker container

```sh
docker compose up -d
```

## Development

### Prerequisites

Install goose and sqlc

**using go:**

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

**using homebrew:**

```sh
brew install goose sqlc
```

### Run database migrations

```sh
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" -dir sql/schema/ up
```
