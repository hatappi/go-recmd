GO := go
ARG = ""

GIT_HASH=$(shell git rev-parse --short HEAD)

GOBIN:=${PWD}/bin
PATH:=${GOBIN}:${PATH}

.PHONY: dependencies
dependencies:
	@${GO} mod download
	@${GO} mod tidy

.PHONY: run
run:
	@${GO} run \
	  -ldflags "-X main.commit=${GIT_HASH}" \
	  cmd/recmd/*.go \
	  ${ARG}

.PHONY: build
build:
	@${GO} build \
	  -ldflags "-X main.commit=${GIT_HASH}" \
	  -o dist/recmd cmd/recmd/*.go

.PHONY: install-tools
install-tools:
	@GOBIN=${GOBIN} ./scripts/install_tools.sh

.PHONY: lint
lint:
	@${GOBIN}/golangci-lint run ./...

.PHONY: test
test:
	@${GO} test ./...
