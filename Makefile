BINARY  := statora
BUILD   := go build -o $(BINARY) .
MODULE  := statora-cli

.PHONY: build run test clean dev lint tidy

## build: Compile the binary
build:
	$(BUILD)

## run: Build and run (pass ARGS="..." for subcommand args)
run: build
	./$(BINARY) $(ARGS)

## test: Run all tests
test:
	go test ./...

## test-verbose: Run all tests with verbose output
test-verbose:
	go test -v ./...

## tidy: Tidy go modules
tidy:
	go mod tidy

## clean: Remove built binary
clean:
	rm -f $(BINARY)

## lint: Run go vet
lint:
	go vet ./...

## dev: Watch source files and rebuild on change (requires watchexec)
##      Install: cargo install watchexec-cli  OR  brew install watchexec
dev:
	watchexec \
		--exts go \
		--restart \
		--clear \
		-- sh -c 'go build -o $(BINARY) . && echo "✓ rebuilt $(BINARY)"'

## install: Install binary to GOPATH/bin
install:
	go install .
