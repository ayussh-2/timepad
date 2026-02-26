BINARY_NAME=timepad

run:
	air

migrate:
	go run cmd/migrate/main.go

seed:
	go run cmd/seed/seed.go

build:
	go build -o tmp/${BINARY_NAME} cmd/main.go

clean:
	go clean
	rm tmp/${BINARY_NAME}