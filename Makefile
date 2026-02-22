GO ?= go
BIN_DIR ?= ./bin

.PHONY: build test test-race test-coverage lint clean
COVERAGE_MIN ?= 10.0

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/screenshot_mcp_server ./cmd/screenshot_mcp_server
	$(GO) build -o $(BIN_DIR)/agent ./cmd/agent

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

test-coverage:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic

test-coverage-gate:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	@coverage=$$(go tool cover -func=coverage.out | awk '/^total:/ {print $$3}' | tr -d '%'); \
	if [ -z "$$coverage" ]; then \
		echo "coverage report missing"; \
		exit 1; \
	fi; \
	if awk -v value="$$coverage" -v min="$(COVERAGE_MIN)" 'BEGIN { exit !(value >= min) }'; then \
		echo "coverage gate passed: $$coverage% (minimum $(COVERAGE_MIN)%)"; \
	else \
		echo "coverage gate failed: $$coverage% (minimum $(COVERAGE_MIN)%)"; \
		exit 1; \
	fi

lint:
	golangci-lint run

clean:
	rm -rf $(BIN_DIR) coverage.out
