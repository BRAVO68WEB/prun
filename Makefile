.PHONY: build test clean install run help

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build the prun binary"
	@echo "  test     - Run tests"
	@echo "  clean    - Remove build artifacts"
	@echo "  install  - Install prun to /usr/local/bin (requires sudo)"
	@echo "  run      - Build and run prun"
	@echo "  help     - Show this help message"

# Build the prun binary
build: clean
	@echo "Building prun..."
	@go build -o prun ./cmd/prun
	@echo "Build complete: ./prun"

# Run tests
test: build
	@echo "Running tests..."
	@chmod +x tests/test.sh
	@tests/test.sh

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f prun
	@go clean

# Install prun to /usr/local/bin (requires sudo)
install: build
	@echo "Installing prun to /usr/local/bin..."
	@sudo mv prun /usr/local/bin/
	@echo "Installation complete"

# Run prun
run: build
	@echo "Running prun..."
	@./prun
