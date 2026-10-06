include .env
export

export PROJECT_ROOT=${shell pwd}

env-up:
	@docker compose up -d murmur-postgres murmur-redis murmur-nats

env-down:
	@docker compose down murmur-postgres murmur-redis murmur-nats

env-cleanup:
	@read -p "Do you plan to clean environments volumes files? Dangerous to lose data. [y:N]: " ans; \
	if [ "$$ans" = "y" ]; then \
	  docker compose down murmur-postgres port-forwarder murmur-nats && \
	  rm -rf out/pgdata && \
	  echo "Environments files are cleaned"; \
	else \
	  echo "Cleaning of environments canceled"; \
	fi

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
  		echo "seq is required param. Example: make migrate-create seq=init"; \
  		exit 1; \
  	fi;
	docker compose run --rm murmur-postgres-migrate \
		create \
		-ext sql \
		--dir /migrations \
		-seq "$(seq)"


migrate-force:
	@docker compose run --rm murmur-postgres-migrate \
		-path /migrations \
		-database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@murmur-postgres:5432/${POSTGRES_DB}?sslmode=disable" \
		force $(version)

migrate-up:
	make migrate-action action=up

migrate-down:
	make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
  		echo "action is required param. Example: make migrate-action action=up"; \
  		exit 1; \
  	fi;
	docker compose run --rm murmur-postgres-migrate \
    		-path /migrations \
    		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@murmur-postgres:5432/${POSTGRES_DB}?sslmode=disable \
    		"$(action)"

murmur-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=localhost && \
	go mod tidy && \
	go run cmd/api/main.go

murmur-run-notification:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=localhost && \
	go mod tidy && \
	go run cmd/notification/main.go

.PHONY: proto
proto:
	@mkdir -p proto/boardpb
	@protoc \
		--go_out=. --go_opt=module=github.com/oleg-morshel/murmur-api \
		--go-grpc_out=. --go-grpc_opt=module=github.com/oleg-morshel/murmur-api \
		proto/board.proto