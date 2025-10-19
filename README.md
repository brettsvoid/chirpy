# Chirpy

## Build and run

```sh
go build -o out && ./out
```

## Run postgres docker container

```sh
docker compose up -d
```

## Run database migrations

```sh
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" -dir sql/schema/ up
```
