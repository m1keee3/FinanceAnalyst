PROTO_DIR := ../../proto/services/scanner/v1
OUT_DIR := proto-gen/v1
PROTOC := protoc

gen:
	@mkdir -p $(OUT_DIR)
	$(PROTOC) \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

clean:
	@rm -rf proto-gen/v1/*.pb.go
