BINARY     := timepad.exe
LDFLAGS    := -H=windowsgui -s -w
ICON_SRC   := ../../icons/logo.png
ICON_DST   := internal/ui/assets/icon.png

export CGO_ENABLED := 1

.PHONY: setup build run clean

setup:
	go mod tidy

$(ICON_DST): $(ICON_SRC)
	powershell -NoProfile -Command "Copy-Item -Force '$(ICON_SRC)' '$(ICON_DST)'"

build: $(ICON_DST)
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

run: $(ICON_DST)
	go run .

clean:
	-powershell -NoProfile -Command "Remove-Item -Force '$(BINARY)' -ErrorAction SilentlyContinue"
	-powershell -NoProfile -Command "Remove-Item -Force '$(ICON_DST)' -ErrorAction SilentlyContinue"
