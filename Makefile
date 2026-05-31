BIN := point-tui

.PHONY: build test clean

build:
	go build -o $(BIN) ./cmd/point-tui

test:
	go test ./...

clean:
	rm -f $(BIN)
