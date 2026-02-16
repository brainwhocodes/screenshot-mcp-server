GO ?= go
BIN_DIR ?= ./bin

.PHONY: build test lint clean

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/screenshot_mcp_server ./cmd/screenshot_mcp_server

test:
	$(GO) test ./...

lint:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR) coverage.out
