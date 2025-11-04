.PHONY: build run install clean test

# Build the application
build:
	go build -o azdo-tui .

# Run the application
run: build
	./azdo-tui

# Install the application to $GOPATH/bin
install:
	go install .

# Clean build artifacts
clean:
	rm -f azdo-tui

# Run tests
test:
	go test ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...
