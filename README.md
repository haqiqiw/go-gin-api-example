# go-gin-api-example

A sample Go API implementation using clean architecture.

This repo is a reworked version of [go-api-example](https://github.com/haqiqiw/go-api-example) by replacing Fiber with Gin.

## Dependencies

- [MySQL](https://www.mysql.com/) – Primary database
- [Redis](https://redis.io/) – Caching and session store
- [Kafka](https://kafka.apache.org/) – Event streaming platform

## Frameworks & Libraries

- [Gin](https://gin-gonic.com/) – HTTP framework for Go
- [godotenv](https://github.com/joho/godotenv) – `.env` file loader
- [golang-migrate](https://github.com/golang-migrate/migrate) – Database migration tool
- [validator](https://github.com/go-playground/validator) – Struct and field validation
- [zap](https://github.com/uber-go/zap) – Structured logging
- [confluent-kafka-go](https://github.com/confluentinc/confluent-kafka-go) – Kafka client for Go
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) – JWT authentication

## Running with Docker Compose

Start all required services:

```bash
docker-compose up -d
```

Stop all services:

```bash
docker-compose down
```

> Make sure all containers are running before running database migration and starting the app.
>
> Kafka UI is available at http://localhost:8080/

## Database Migration

Apply migration:

```bash
migrate -path ./db/migrations -database "mysql://root:root@tcp(localhost:3306)/api-example" up
```

Rollback migration:

```bash
migrate -path ./db/migrations -database "mysql://root:root@tcp(localhost:3306)/api-example" down
```

Create new migration file:

```bash
migrate create -ext sql -dir db/migrations <migration_name>
```

## Running the Application

Copy `.env` from `env.sample` and adjust configuration values:

```bash
cp env.sample .env
```

Run the API server:

```bash
go run cmd/api/main.go
```

Run the Kafka consumer:

```bash
go run cmd/consumer/main.go
```
