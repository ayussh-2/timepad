BINARY_NAME=timepad

.PHONY: dev migrate seed build clean test test-v test-run test-race vet lint purge

dev:
	air

migrate:
	go run cmd/migrate/main.go

seed:
	go run cmd/seed/seed.go

purge:
	go run cmd/purge/main.go

build:
	go build -o tmp/${BINARY_NAME} cmd/main.go

clean:
	go clean
	rm -f tmp/${BINARY_NAME}

# ── Testing ────────────────────────────────────────────────────────────────────

## Run all tests (silent output)
test:
	go test ./tests/... -timeout 60s

## Run all tests with verbose output
test-v:
	go test ./tests/... -v -timeout 60s

## Run a specific test by name  →  make test-run NAME=TestSettings
test-run:
	go test ./tests/... -v -run $(NAME) -timeout 60s

## Run all tests with the race detector
test-race:
	go test ./tests/... -v -race -timeout 120s

# ── Code quality ───────────────────────────────────────────────────────────────

vet:
	go vet ./...
