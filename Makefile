.DEFAULT_GOAL := build

.PHONY: build
build:
	go build -o ai-commit


.PHONY: clean
clean:
	go clean
	rm -f ci-commit

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build   Build the ci-commit binary"
	@echo "  test    Run tests"
	@echo "  clean   Clean up the project"
	@echo "  lint    Run linters"
	@echo "  fmt     Format source code"
	@echo "  help    Show this message"
