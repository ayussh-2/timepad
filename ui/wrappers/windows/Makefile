BINARY  := timepad.exe
LDFLAGS := -H=windowsgui -s -w

export CGO_ENABLED := 1

.PHONY: setup build run clean

setup:
	go mod tidy

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

run:
	go run .

clean:
	-rm -f $(BINARY)
