BINARY := qrl
PKG := ./cmd/qrl
E2E_TAG := e2e

.PHONY: build run test e2e clean

build:
	go build -o $(BINARY) $(PKG)

run: build
	./$(BINARY)

test: build
	go test ./...

e2e: build
	go test -tags=$(E2E_TAG) ./e2e -v

clean:
	rm -f $(BINARY)
