# Makefile
#
# Why we chose a Makefile approach:
# Make provides a simple, standard way for developers to run consistent
# build and generation commands with a single command (e.g., "make proto").
# It avoids duplication across different scripts and ensures everyone
# uses the same steps.

PROTO_DIR := ./proto
PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto")

.PHONY: proto
proto:
	@protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_DIR) \
		--go-grpc_out=$(PROTO_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
	@echo "Protobuf generation complete!"

.PHONY: test
test:
    @go test ./...
    @echo "Go tests complete!"
