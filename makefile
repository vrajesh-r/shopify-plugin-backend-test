all: run_be run_fe run_local db_boot
.PHONY: all
# following commands are for running locally (no docker, no slice)
run_local: db_boot run_be run_fe
run_fe:
	#opens in new tab
	cd ../shopify_plugin_frontend && npm run build && npm run dev
run_be:
	ENVIRONMENT=localdev \
	CONFIG_FILE_NAME=config \
	CONFIG_FILE_PATH=../deploy/chart/ \
	cd service && go run ./cmd/shopify_plugin_backend
run_be_env:
	IS_CLASSIC=true \
	SERVE_FE_ROUTES=true \
	cd service && go run ./cmd/shopify_plugin_backend
run_be_config:
	IS_CLASSIC=false \
	cd service && go run ./cmd/shopify_plugin_backend
db_boot:
	docker-compose up -d redis db
	sleep 5
	./scripts/build-local-db.sh
	echo "applying goose migrations (honk honk)"
	cd service/internal/storage/migrations/postgres/migrations && \
	goose -path .. up
	#goose -env stagingintegrations "dbname=shopify sslmode=disable" up
### below are unused in current iteration
rebuild:
	docker-compose up --build
#local dev only
install:
	scripts/bootstrap.sh
connect_db:	
	ifneq (,$(wildcard ./.env))
		include .env
		export
	endif
	psql postgresql://${DB_USER}:${DB_PASSWORD}@localhost:${DB_EXPOSED_PORT}
phone:
	scripts/gen_phone_number.sh
test:
	go test ./...
build-debug:
	DOCKERFILE="scripts/Dockerfile-debug" docker-compose build 
	echo "debug docker image built. execute 'make run' to launch and then 'make debug' in separate window to debug"
debug:
#requires debug build 
	dlv connect localhost:40000
run-no-docker:
	docker-compose up -d db redis
	./scripts/milton-local.sh