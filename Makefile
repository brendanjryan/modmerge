.PHONY: default
default: build

.PHONY: build ## Builds and ensures all of the source files in this repo are valid.
build:
	go build -o bin/modmerge

.PHONY: test ## Runs unit tests for this repo.
test:
	go test ./...
