PROTOC := protoc
OUT_DIR := proto-gen

WATCHER_PROTO_DIR := ../../proto/services/watcher/v1
SCANNER_PROTO_DIR := ../../proto/services/scanner/v1

MIGRATOR := go run ./cmd/migrator
MIGRATIONS_DIR := ./migrations

gen-all: gen-watcher gen-scanner

gen-watcher:
	@mkdir -p $(OUT_DIR)/watcher/v1
	$(PROTOC) \
		--proto_path=$(WATCHER_PROTO_DIR) \
		--go_out=$(OUT_DIR)/watcher/v1 --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR)/watcher/v1 --go-grpc_opt=paths=source_relative \
		$(WATCHER_PROTO_DIR)/*.proto

gen-scanner:
	@mkdir -p $(OUT_DIR)/scanner/v1
	$(PROTOC) \
		--proto_path=$(SCANNER_PROTO_DIR) \
		--go_out=$(OUT_DIR)/scanner/v1 --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR)/scanner/v1 --go-grpc_opt=paths=source_relative \
		$(SCANNER_PROTO_DIR)/*.proto

clean:
	@rm -rf $(OUT_DIR)/*

migrate-up:
	@test -n "$(POSTGRES_DSN)" || (echo "POSTGRES_DSN is not set" && exit 1)
	$(MIGRATOR) \
		-dsn "$(POSTGRES_DSN)" \
		-migrations-path "$(MIGRATIONS_DIR)"

migrate-down:
	@test -n "$(POSTGRES_DSN)" || (echo "POSTGRES_DSN is not set" && exit 1)
	$(MIGRATOR) \
		-dsn "$(POSTGRES_DSN)" \
		-migrations-path "$(MIGRATIONS_DIR)" \
		-down