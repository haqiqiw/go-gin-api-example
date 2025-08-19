MIGRATE = migrate
MIGRATION_DIR = ./db/migrations
DB_URL = "mysql://root:root@tcp(localhost:3306)/api-example"

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	$(MIGRATE) -path $(MIGRATION_DIR) -database $(DB_URL) up

migrate-down:
	$(MIGRATE) -path $(MIGRATION_DIR) -database $(DB_URL) down

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=<migration_name>"; \
		exit 1; \
	fi
	$(MIGRATE) create -ext sql -dir $(MIGRATION_DIR) $(name)

env:
	cp env.sample .env

run-api:
	go run cmd/api/main.go

run-consumer:
	go run cmd/consumer/main.go