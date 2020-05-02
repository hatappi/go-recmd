GO := go

.PHONY: run
run:
	@${GO} run cmd/recmd/main.go

.PHONY: build
build:
	@${GO} build -o dist/recmd cmd/recmd/main.go
