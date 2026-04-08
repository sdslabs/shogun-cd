DB_CONTAINER_NAME=shogun_db
DB_PORT=5432
DB_USER=shogun
DB_PASSWORD=pwd
DB_NAME=shogun_db

.PHONY: db-run db-down db-logs start

start:
	make db-run || true
	@echo "Starting Shogun-CD"
	air

db-run:
	@echo "Starting database container '$(DB_CONTAINER_NAME)' on port $(DB_PORT)..."
	docker run --name $(DB_CONTAINER_NAME) \
		--rm \
		-v $(PWD)/.data/db_volume:/var/lib/postgresql/ \
		-p $(DB_PORT):5432 \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-d postgres:18.3-alpine3.23
	@echo "Database is ready!"

db-down:
	@echo "Stopping database..."
	docker stop $(DB_CONTAINER_NAME)
	@echo "Database stopped."

db-logs:
	docker logs -f $(DB_CONTAINER_NAME)