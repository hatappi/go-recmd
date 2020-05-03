GO := go

GOBIN:=${PWD}/bin
PATH:=${GOBIN}:${PATH}

.PHONY: dependencies
dependencies:
	go mod download
	go mod tidy

.PHONY: run
run:
	@${GO} run cmd/recmd/main.go

.PHONY: build
build:
	@${GO} build -o dist/recmd cmd/recmd/main.go

.PHONY: install-tools
install-tools:
	@GOBIN=${GOBIN} ./scripts/install_tools.sh

.PHONY: lint
lint:
	@${GOBIN}/golangci-lint run ./...

.PHONY: test
test:
	@${GO} test ./...
