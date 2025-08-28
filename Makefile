.PHONY: generate-proto clean

# Директория с .proto файлами
PROTO_DIR := proto

# Команда для генерации Go-кода из .proto файлов
generate-proto:
	@echo "Generating Go code from protobuf files..."
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/receiver.proto $(PROTO_DIR)/service.proto
	@echo "Done."

clean:
	@echo "Cleaning generated protobuf files..."
	rm -f $(PROTO_DIR)/*.pb.go
	@echo "Done."
